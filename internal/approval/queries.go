package approval

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"greenpost/db/generated"
	"greenpost/internal/identity"
	"greenpost/internal/platform/pgxutil"

	"github.com/jackc/pgx/v5"
)

type ApprovalRequestSummary struct {
	ID                string                `json:"id"`
	WorkspaceID       string                `json:"workspace_id"`
	ActionType        string                `json:"action_type"`
	Title             string                `json:"title"`
	Description       string                `json:"description,omitempty"`
	Status            string                `json:"status"`
	DecisionRequired  bool                  `json:"decision_required"`
	MatchedPolicy     *MatchedPolicySummary `json:"matched_policy,omitempty"`
	SourcePlatform    string                `json:"source_platform"`
	SourceWorkflowID  string                `json:"source_workflow_id"`
	SourceExecutionID string                `json:"source_execution_id"`
	OriginalAction    json.RawMessage       `json:"original_action"`
	Context           json.RawMessage       `json:"context"`
	CreatedAt         string                `json:"created_at"`
	ResolvedAt        string                `json:"resolved_at,omitempty"`
	ExpiresAt         string                `json:"expires_at,omitempty"`
}

type MatchedPolicySummary struct {
	ID              string `json:"id"`
	VersionID       string `json:"version_id"`
	Name            string `json:"name"`
	Effect          string `json:"effect"`
	Priority        int32  `json:"priority"`
	VersionNumber   int32  `json:"version_number"`
	DeadlineSeconds int64  `json:"deadline_seconds"`
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

type AuditEventSummary struct {
	ID                string          `json:"id"`
	WorkspaceID       string          `json:"workspace_id"`
	ApprovalRequestID string          `json:"approval_request_id"`
	DecisionID        string          `json:"decision_id,omitempty"`
	ActorType         string          `json:"actor_type"`
	ActorID           string          `json:"actor_id,omitempty"`
	EventType         string          `json:"event_type"`
	Metadata          json.RawMessage `json:"metadata"`
	CreatedAt         string          `json:"created_at"`
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

type ListApprovalRequestAuditEventsCommand struct {
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

func (s *Service) ListApprovalRequestAuditEvents(ctx context.Context, cmd ListApprovalRequestAuditEventsCommand) ([]AuditEventSummary, error) {
	if strings.TrimSpace(cmd.WorkspaceID) == "" || strings.TrimSpace(cmd.RequestID) == "" || strings.TrimSpace(cmd.UserID) == "" {
		return nil, ErrInvalidRequest
	}

	workspaceID, err := pgxutil.UUIDText(cmd.WorkspaceID)
	if err != nil {
		return nil, identity.ErrInvalidWorkspaceID
	}

	requestID, err := pgxutil.UUIDText(cmd.RequestID)
	if err != nil {
		return nil, ErrInvalidRequest
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

	if _, err := queries.GetApprovalRequestByID(ctx, generated.GetApprovalRequestByIDParams{
		WorkspaceID: workspaceID,
		ID:          requestID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("load approval request: %w", err)
	}

	events, err := queries.ListAuditEventsByRequestID(ctx, generated.ListAuditEventsByRequestIDParams{
		WorkspaceID:       workspaceID,
		ApprovalRequestID: requestID,
	})
	if err != nil {
		return nil, fmt.Errorf("list approval request audit events: %w", err)
	}

	summaries := make([]AuditEventSummary, 0, len(events))
	for _, event := range events {
		summaries = append(summaries, auditEventSummary(event))
	}

	return summaries, nil
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
		MatchedPolicy:     matchedPolicySummary(request.MatchedPolicySnapshot),
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

func matchedPolicySummary(snapshot []byte) *MatchedPolicySummary {
	if len(bytes.TrimSpace(snapshot)) == 0 {
		return nil
	}

	var parsed struct {
		PolicyID         string          `json:"policy_id"`
		PolicyVersionID  string          `json:"policy_version_id"`
		Name             string          `json:"name"`
		Priority         int32           `json:"priority"`
		VersionNumber    int32           `json:"version_number"`
		Effect           string          `json:"effect"`
		ApprovalSettings json.RawMessage `json:"approval_settings"`
	}
	if err := json.Unmarshal(snapshot, &parsed); err != nil {
		return nil
	}

	policyID := strings.TrimSpace(parsed.PolicyID)
	versionID := strings.TrimSpace(parsed.PolicyVersionID)
	if policyID == "" || versionID == "" {
		return nil
	}

	summary := &MatchedPolicySummary{
		ID:            policyID,
		VersionID:     versionID,
		Name:          parsed.Name,
		Effect:        parsed.Effect,
		Priority:      parsed.Priority,
		VersionNumber: parsed.VersionNumber,
	}

	if len(bytes.TrimSpace(parsed.ApprovalSettings)) > 0 && !bytes.Equal(bytes.TrimSpace(parsed.ApprovalSettings), []byte("null")) {
		var settings struct {
			DeadlineSeconds int64 `json:"deadline_seconds"`
		}
		if err := json.Unmarshal(parsed.ApprovalSettings, &settings); err == nil {
			summary.DeadlineSeconds = settings.DeadlineSeconds
		}
	}

	return summary
}

func auditEventSummary(event generated.AuditEvent) AuditEventSummary {
	return AuditEventSummary{
		ID:                pgxutil.UUIDString(event.ID),
		WorkspaceID:       pgxutil.UUIDString(event.WorkspaceID),
		ApprovalRequestID: pgxutil.UUIDString(event.ApprovalRequestID),
		DecisionID:        pgxutil.UUIDString(event.DecisionID),
		ActorType:         event.ActorType,
		ActorID:           pgxutil.UUIDString(event.ActorID),
		EventType:         event.EventType,
		Metadata:          json.RawMessage(event.Metadata),
		CreatedAt:         event.CreatedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
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
