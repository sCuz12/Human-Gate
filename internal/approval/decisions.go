package approval

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"greenpost/db/generated"
	"greenpost/internal/identity"
	"greenpost/internal/platform/pgxutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type DecisionCommand struct {
	WorkspaceID string
	RequestID   string
	UserID      string
	Decision    string
	Comment     string
}

func (s *Service) DecideApprovalRequest(ctx context.Context, cmd DecisionCommand) (ApprovalRequestSummary, error) {
	if strings.TrimSpace(cmd.WorkspaceID) == "" || strings.TrimSpace(cmd.RequestID) == "" || strings.TrimSpace(cmd.UserID) == "" {
		return ApprovalRequestSummary{}, ErrInvalidRequest
	}

	workspaceID, err := pgxutil.UUIDText(cmd.WorkspaceID)
	if err != nil {
		return ApprovalRequestSummary{}, identity.ErrInvalidWorkspaceID
	}

	requestID, err := pgxutil.UUIDText(cmd.RequestID)
	if err != nil {
		return ApprovalRequestSummary{}, ErrInvalidRequest
	}

	userID, err := pgxutil.UUIDText(cmd.UserID)
	if err != nil {
		return ApprovalRequestSummary{}, identity.ErrInvalidUserID
	}

	decision, resolvedStatus, eventType, err := normalizeDecision(cmd.Decision)
	if err != nil {
		return ApprovalRequestSummary{}, err
	}

	s.logDecisionStep(ctx, cmd, "begin_transaction")
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ApprovalRequestSummary{}, fmt.Errorf("begin decision transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := generated.New(tx)
	s.logDecisionStep(ctx, cmd, "check_workspace_membership")
	if _, err := queries.GetWorkspaceMember(ctx, generated.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ApprovalRequestSummary{}, ErrForbidden
		}
		return ApprovalRequestSummary{}, fmt.Errorf("load workspace member: %w", err)
	}

	s.logDecisionStep(ctx, cmd, "lock_approval_request")
	request, err := queries.LockApprovalRequestForDecision(ctx, generated.LockApprovalRequestForDecisionParams{
		WorkspaceID: workspaceID,
		ID:          requestID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ApprovalRequestSummary{}, ErrNotFound
		}
		return ApprovalRequestSummary{}, fmt.Errorf("lock approval request: %w", err)
	}

	if request.Status != generated.ApprovalStatusPending {
		return ApprovalRequestSummary{}, ErrResolved
	}

	now := s.clock().UTC()
	if request.ExpiresAt.Valid && !request.ExpiresAt.Time.After(now) {
		return ApprovalRequestSummary{}, ErrExpired
	}

	approvedAction := string(request.OriginalAction)
	approvedActionHash := pgxutil.Text(request.OriginalActionHash)
	if decision == generated.DecisionTypeRejected {
		approvedAction = `null`
		approvedActionHash = pgtype.Text{}
	}

	s.logDecisionStep(ctx, cmd, "create_approval_decision")
	createdDecision, err := queries.CreateApprovalDecision(ctx, generated.CreateApprovalDecisionParams{
		WorkspaceID:        workspaceID,
		ApprovalRequestID:  request.ID,
		Decision:           decision,
		OriginalActionHash: request.OriginalActionHash,
		ApprovedAction:     approvedAction,
		ApprovedActionHash: approvedActionHash,
		ChangedFields:      `[]`,
		Comment:            pgxutil.Text(cmd.Comment),
		DecidedBy:          userID,
		IssuedAt:           pgxutil.Timestamptz(now),
		ExpiresAt:          pgtype.Timestamptz{},
	})
	if err != nil {
		return ApprovalRequestSummary{}, fmt.Errorf("create approval decision: %w", err)
	}

	s.logDecisionStep(ctx, cmd, "load_continuation_target")
	continuationTarget, err := queries.GetContinuationTargetByRequestID(ctx, generated.GetContinuationTargetByRequestIDParams{
		WorkspaceID:       workspaceID,
		ApprovalRequestID: request.ID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ApprovalRequestSummary{}, fmt.Errorf("load continuation target: %w", ErrNotFound)
		}
		return ApprovalRequestSummary{}, fmt.Errorf("load continuation target: %w", err)
	}

	s.logDecisionStep(ctx, cmd, "create_decision_delivery")
	createdDelivery, err := queries.CreateDecisionDelivery(ctx, generated.CreateDecisionDeliveryParams{
		WorkspaceID:          workspaceID,
		DecisionID:           createdDecision.ID,
		ContinuationTargetID: continuationTarget.ID,
		Status:               generated.DeliveryStatusPending,
		NextAttemptAt:        pgxutil.Timestamptz(now),
	})
	if err != nil {
		return ApprovalRequestSummary{}, fmt.Errorf("create decision delivery: %w", err)
	}

	s.logDecisionStep(ctx, cmd, "resolve_approval_request")
	resolved, err := queries.ResolveApprovalRequest(ctx, generated.ResolveApprovalRequestParams{
		WorkspaceID: workspaceID,
		ID:          request.ID,
		Status:      resolvedStatus,
		ResolvedAt:  pgxutil.Timestamptz(now),
	})
	if err != nil {
		return ApprovalRequestSummary{}, fmt.Errorf("resolve approval request: %w", err)
	}

	auditMetadata, err := marshalJSONObject(map[string]any{
		"decision": decision,
		"comment":  strings.TrimSpace(cmd.Comment) != "",
	})
	if err != nil {
		return ApprovalRequestSummary{}, fmt.Errorf("marshal audit metadata: %w", err)
	}

	s.logDecisionStep(ctx, cmd, "create_audit_event")
	if _, err := queries.CreateAuditEvent(ctx, generated.CreateAuditEventParams{
		WorkspaceID:       workspaceID,
		ApprovalRequestID: request.ID,
		DecisionID:        createdDecision.ID,
		ActorType:         "user",
		ActorID:           userID,
		EventType:         eventType,
		Metadata:          string(auditMetadata),
	}); err != nil {
		return ApprovalRequestSummary{}, fmt.Errorf("create decision audit event: %w", err)
	}

	deliveryAuditMetadata, err := marshalJSONObject(map[string]any{
		"delivery_id":              pgxutil.UUIDString(createdDelivery.ID),
		"continuation_target_id":   pgxutil.UUIDString(continuationTarget.ID),
		"delivery_status":          createdDelivery.Status,
		"continuation_strategy":    continuationTarget.Strategy,
		"continuation_platform":    continuationTarget.Platform,
		"delivery_next_attempt_at": now,
	})
	if err != nil {
		return ApprovalRequestSummary{}, fmt.Errorf("marshal delivery audit metadata: %w", err)
	}

	s.logDecisionStep(ctx, cmd, "create_delivery_audit_event")
	if _, err := queries.CreateAuditEvent(ctx, generated.CreateAuditEventParams{
		WorkspaceID:       workspaceID,
		ApprovalRequestID: request.ID,
		DecisionID:        createdDecision.ID,
		ActorType:         "system",
		ActorID:           pgtype.UUID{},
		EventType:         "decision.delivery_scheduled",
		Metadata:          string(deliveryAuditMetadata),
	}); err != nil {
		return ApprovalRequestSummary{}, fmt.Errorf("create delivery audit event: %w", err)
	}

	s.logDecisionStep(ctx, cmd, "commit_transaction")
	if err := tx.Commit(ctx); err != nil {
		return ApprovalRequestSummary{}, fmt.Errorf("commit decision transaction: %w", err)
	}

	s.logDecisionStep(ctx, cmd, "decision_completed")
	return approvalRequestSummary(resolved), nil
}

func (s *Service) logDecisionStep(ctx context.Context, cmd DecisionCommand, step string) {
	s.logger.InfoContext(
		ctx,
		"approval decision step",
		"step", step,
		"workspace_id", cmd.WorkspaceID,
		"approval_request_id", cmd.RequestID,
		"user_id", cmd.UserID,
		"decision", cmd.Decision,
	)
}

func normalizeDecision(decision string) (generated.DecisionType, generated.ApprovalStatus, string, error) {
	switch decision {
	case "approve":
		return generated.DecisionTypeApproved, generated.ApprovalStatusApproved, "approval_request.approved", nil
	case "reject":
		return generated.DecisionTypeRejected, generated.ApprovalStatusRejected, "approval_request.rejected", nil
	default:
		return "", "", "", ErrInvalidRequest
	}
}
