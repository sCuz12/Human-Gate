package approval

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"greenpost/db/generated"
	"greenpost/internal/platform/pgxutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const defaultExpiryBatchSize int32 = 25

type ExpiryResult struct {
	Expired int
}

func (s *Service) ExpireDueApprovalRequests(ctx context.Context, limit int32) (ExpiryResult, error) {
	if limit <= 0 {
		limit = defaultExpiryBatchSize
	}

	now := s.clock().UTC()
	queries := generated.New(s.db)

	requests, err := queries.ListDueExpiredApprovalRequests(ctx, generated.ListDueExpiredApprovalRequestsParams{
		NowAt:      pgxutil.Timestamptz(now),
		LimitCount: limit,
	})
	if err != nil {
		return ExpiryResult{}, fmt.Errorf("list due expired approval requests: %w", err)
	}

	result := ExpiryResult{}
	for _, request := range requests {
		expired, err := s.expireOne(ctx, request.WorkspaceID, request.ID, now)
		if err != nil {
			s.logger.ErrorContext(
				ctx,
				"expire approval request failed",
				"error", err,
				"workspace_id", pgxutil.UUIDString(request.WorkspaceID),
				"approval_request_id", pgxutil.UUIDString(request.ID),
			)
			continue
		}
		if expired {
			result.Expired++
		}
	}

	if result.Expired > 0 {
		s.logger.InfoContext(ctx, "approval requests expired", slog.Int("count", result.Expired))
	}

	return result, nil
}

func (s *Service) expireOne(ctx context.Context, workspaceID pgtype.UUID, requestID pgtype.UUID, now time.Time) (bool, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, fmt.Errorf("begin expiry transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := generated.New(tx)
	request, err := queries.LockApprovalRequestForDecision(ctx, generated.LockApprovalRequestForDecisionParams{
		WorkspaceID: workspaceID,
		ID:          requestID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("lock approval request for expiry: %w", err)
	}

	if request.Status != generated.ApprovalStatusPending {
		return false, nil
	}
	if !request.ExpiresAt.Valid || request.ExpiresAt.Time.After(now) {
		return false, nil
	}

	createdDecision, err := queries.CreateApprovalDecision(ctx, generated.CreateApprovalDecisionParams{
		WorkspaceID:        workspaceID,
		ApprovalRequestID:  request.ID,
		Decision:           generated.DecisionTypeExpired,
		OriginalActionHash: request.OriginalActionHash,
		ApprovedAction:     `null`,
		ApprovedActionHash: pgtype.Text{},
		ChangedFields:      `[]`,
		Comment:            pgxutil.Text("Approval request expired before a decision was recorded."),
		DecidedBy:          pgtype.UUID{},
		IssuedAt:           pgxutil.Timestamptz(now),
		ExpiresAt:          pgtype.Timestamptz{},
	})
	if err != nil {
		return false, fmt.Errorf("create expired approval decision: %w", err)
	}

	continuationTarget, err := queries.GetContinuationTargetByRequestID(ctx, generated.GetContinuationTargetByRequestIDParams{
		WorkspaceID:       workspaceID,
		ApprovalRequestID: request.ID,
	})
	if err != nil {
		return false, fmt.Errorf("load expiry continuation target: %w", err)
	}

	createdDelivery, err := queries.CreateDecisionDelivery(ctx, generated.CreateDecisionDeliveryParams{
		WorkspaceID:          workspaceID,
		DecisionID:           createdDecision.ID,
		ContinuationTargetID: continuationTarget.ID,
		Status:               generated.DeliveryStatusPending,
		NextAttemptAt:        pgxutil.Timestamptz(now),
	})
	if err != nil {
		return false, fmt.Errorf("create expired decision delivery: %w", err)
	}

	if _, err := queries.ResolveApprovalRequest(ctx, generated.ResolveApprovalRequestParams{
		WorkspaceID: workspaceID,
		ID:          request.ID,
		Status:      generated.ApprovalStatusExpired,
		ResolvedAt:  pgxutil.Timestamptz(now),
	}); err != nil {
		return false, fmt.Errorf("resolve approval request expired: %w", err)
	}

	auditMetadata, err := marshalJSONObject(map[string]any{
		"expired_at": now,
		"expires_at": request.ExpiresAt.Time.UTC(),
	})
	if err != nil {
		return false, fmt.Errorf("marshal expiry audit metadata: %w", err)
	}

	if _, err := queries.CreateAuditEvent(ctx, generated.CreateAuditEventParams{
		WorkspaceID:       workspaceID,
		ApprovalRequestID: request.ID,
		DecisionID:        createdDecision.ID,
		ActorType:         "system",
		ActorID:           pgtype.UUID{},
		EventType:         "approval_request.expired",
		Metadata:          string(auditMetadata),
	}); err != nil {
		return false, fmt.Errorf("create expiry audit event: %w", err)
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
		return false, fmt.Errorf("marshal expiry delivery audit metadata: %w", err)
	}

	if _, err := queries.CreateAuditEvent(ctx, generated.CreateAuditEventParams{
		WorkspaceID:       workspaceID,
		ApprovalRequestID: request.ID,
		DecisionID:        createdDecision.ID,
		ActorType:         "system",
		ActorID:           pgtype.UUID{},
		EventType:         "decision.delivery_scheduled",
		Metadata:          string(deliveryAuditMetadata),
	}); err != nil {
		return false, fmt.Errorf("create expiry delivery audit event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit expiry transaction: %w", err)
	}

	return true, nil
}
