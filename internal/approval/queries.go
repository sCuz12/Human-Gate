package approval

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"humangate/db/generated"
	"humangate/internal/identity"
	"humangate/internal/platform/pgxutil"

	"github.com/jackc/pgx/v5"
)

type ApprovalRequestSummary struct {
	ID                string          `json:"id"`
	WorkspaceID       string          `json:"workspace_id"`
	ActionType        string          `json:"action_type"`
	Title             string          `json:"title"`
	Description       string          `json:"description,omitempty"`
	Status            string          `json:"status"`
	DecisionRequired  bool            `json:"decision_required"`
	SourcePlatform    string          `json:"source_platform"`
	SourceWorkflowID  string          `json:"source_workflow_id"`
	SourceExecutionID string          `json:"source_execution_id"`
	OriginalAction    json.RawMessage `json:"original_action"`
	Context           json.RawMessage `json:"context"`
	CreatedAt         string          `json:"created_at"`
	ResolvedAt        string          `json:"resolved_at,omitempty"`
	ExpiresAt         string          `json:"expires_at,omitempty"`
}

type DeliverySummary struct {
	ID               string `json:"id"`
	DecisionID       string `json:"decision_id"`
	Status           string `json:"status"`
	AttemptCount     int32  `json:"attempt_count"`
	NextAttemptAt    string `json:"next_attempt_at,omitempty"`
	LastAttemptAt    string `json:"last_attempt_at,omitempty"`
	LastResponseCode int32  `json:"last_response_code,omitempty"`
	LastError        string `json:"last_error,omitempty"`
	DeliveredAt      string `json:"delivered_at,omitempty"`
	AcknowledgedAt   string `json:"acknowledged_at,omitempty"`
	UpdatedAt        string `json:"updated_at"`
}

type ListApprovalRequestsCommand struct {
	WorkspaceID string
	UserID      string
	Limit       int32
	Offset      int32
}

type GetApprovalRequestCommand struct {
	WorkspaceID string
	RequestID   string
	UserID      string
}

type GetApprovalRequestDeliveryCommand struct {
	WorkspaceID string
	RequestID   string
	UserID      string
}

func (s *Service) GetApprovalRequest(ctx context.Context, cmd GetApprovalRequestCommand) (ApprovalRequestSummary, error) {
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

	queries := generated.New(s.db)
	if _, err := queries.GetWorkspaceMember(ctx, generated.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ApprovalRequestSummary{}, ErrForbidden
		}
		return ApprovalRequestSummary{}, fmt.Errorf("load workspace member: %w", err)
	}

	request, err := queries.GetApprovalRequestByID(ctx, generated.GetApprovalRequestByIDParams{
		WorkspaceID: workspaceID,
		ID:          requestID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ApprovalRequestSummary{}, ErrNotFound
		}
		return ApprovalRequestSummary{}, fmt.Errorf("load approval request: %w", err)
	}

	return approvalRequestSummary(request), nil
}

func (s *Service) GetApprovalRequestDelivery(ctx context.Context, cmd GetApprovalRequestDeliveryCommand) (DeliverySummary, error) {
	if strings.TrimSpace(cmd.WorkspaceID) == "" || strings.TrimSpace(cmd.RequestID) == "" || strings.TrimSpace(cmd.UserID) == "" {
		return DeliverySummary{}, ErrInvalidRequest
	}

	workspaceID, err := pgxutil.UUIDText(cmd.WorkspaceID)
	if err != nil {
		return DeliverySummary{}, identity.ErrInvalidWorkspaceID
	}

	requestID, err := pgxutil.UUIDText(cmd.RequestID)
	if err != nil {
		return DeliverySummary{}, ErrInvalidRequest
	}

	userID, err := pgxutil.UUIDText(cmd.UserID)
	if err != nil {
		return DeliverySummary{}, identity.ErrInvalidUserID
	}

	queries := generated.New(s.db)
	if _, err := queries.GetWorkspaceMember(ctx, generated.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DeliverySummary{}, ErrForbidden
		}
		return DeliverySummary{}, fmt.Errorf("load workspace member: %w", err)
	}

	delivery, err := queries.GetDecisionDeliveryByApprovalRequestID(ctx, generated.GetDecisionDeliveryByApprovalRequestIDParams{
		WorkspaceID:       workspaceID,
		ApprovalRequestID: requestID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DeliverySummary{}, ErrNotFound
		}
		return DeliverySummary{}, fmt.Errorf("load decision delivery: %w", err)
	}

	return deliverySummary(delivery), nil
}

func (s *Service) ListApprovalRequests(ctx context.Context, cmd ListApprovalRequestsCommand) ([]ApprovalRequestSummary, error) {
	if strings.TrimSpace(cmd.WorkspaceID) == "" || strings.TrimSpace(cmd.UserID) == "" {
		return nil, ErrInvalidRequest
	}

	workspaceID, err := pgxutil.UUIDText(cmd.WorkspaceID)
	if err != nil {
		return nil, identity.ErrInvalidWorkspaceID
	}

	userID, err := pgxutil.UUIDText(cmd.UserID)
	if err != nil {
		return nil, identity.ErrInvalidUserID
	}

	queries := generated.New(s.db)
	if _, err := queries.GetWorkspaceMember(ctx, generated.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrForbidden
		}
		return nil, fmt.Errorf("load workspace member: %w", err)
	}

	limit := cmd.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	requests, err := queries.ListApprovalRequestsByWorkspace(ctx, generated.ListApprovalRequestsByWorkspaceParams{
		WorkspaceID: workspaceID,
		Limit:       limit,
		Offset:      max(cmd.Offset, 0),
	})
	if err != nil {
		return nil, fmt.Errorf("list approval requests: %w", err)
	}

	summaries := make([]ApprovalRequestSummary, 0, len(requests))
	for _, request := range requests {
		summaries = append(summaries, approvalRequestSummary(request))
	}

	return summaries, nil
}

func approvalRequestSummary(request generated.ApprovalRequest) ApprovalRequestSummary {
	summary := ApprovalRequestSummary{
		ID:                pgxutil.UUIDString(request.ID),
		WorkspaceID:       pgxutil.UUIDString(request.WorkspaceID),
		ActionType:        request.ActionType,
		Title:             request.Title,
		Description:       request.Description.String,
		Status:            string(request.Status),
		DecisionRequired:  request.DecisionRequired,
		SourcePlatform:    request.SourcePlatform,
		SourceWorkflowID:  request.SourceWorkflowID,
		SourceExecutionID: request.SourceExecutionID,
		OriginalAction:    json.RawMessage(request.OriginalAction),
		Context:           json.RawMessage(request.Context),
		CreatedAt:         request.CreatedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}

	if request.ResolvedAt.Valid {
		summary.ResolvedAt = request.ResolvedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	if request.ExpiresAt.Valid {
		summary.ExpiresAt = request.ExpiresAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
	}

	return summary
}

func deliverySummary(delivery generated.DecisionDelivery) DeliverySummary {
	summary := DeliverySummary{
		ID:           pgxutil.UUIDString(delivery.ID),
		DecisionID:   pgxutil.UUIDString(delivery.DecisionID),
		Status:       string(delivery.Status),
		AttemptCount: delivery.AttemptCount,
		UpdatedAt:    delivery.UpdatedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}

	if delivery.NextAttemptAt.Valid {
		summary.NextAttemptAt = delivery.NextAttemptAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	if delivery.LastAttemptAt.Valid {
		summary.LastAttemptAt = delivery.LastAttemptAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	if delivery.LastResponseCode.Valid {
		summary.LastResponseCode = delivery.LastResponseCode.Int32
	}
	if delivery.LastError.Valid {
		summary.LastError = delivery.LastError.String
	}
	if delivery.DeliveredAt.Valid {
		summary.DeliveredAt = delivery.DeliveredAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	if delivery.AcknowledgedAt.Valid {
		summary.AcknowledgedAt = delivery.AcknowledgedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
	}

	return summary
}
