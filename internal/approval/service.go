package approval

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"greenpost/db/generated"
	"greenpost/internal/platform/pgxutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Clock func() time.Time

type Service struct {
	db     *pgxpool.Pool
	clock  Clock
	logger *slog.Logger
}

type SubmitResult struct {
	Request  generated.ApprovalRequest
	Existing bool
}

func NewService(db *pgxpool.Pool, clock Clock, loggers ...*slog.Logger) *Service {
	if clock == nil {
		clock = time.Now
	}

	logger := slog.Default()
	if len(loggers) > 0 && loggers[0] != nil {
		logger = loggers[0]
	}

	return &Service{
		db:     db,
		clock:  clock,
		logger: logger,
	}
}

func (s *Service) SubmitApprovalRequest(ctx context.Context, cmd SubmitApprovalRequestCommand) (SubmitResult, error) {
	if err := validateSubmitCommand(cmd); err != nil {
		return SubmitResult{}, err
	}

	workspaceID, err := pgxutil.UUIDText(cmd.WorkspaceID)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("parse workspace id: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return SubmitResult{}, fmt.Errorf("begin approval request transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := generated.New(tx)

	existing, err := queries.GetApprovalRequestByIdempotencyKey(ctx, generated.GetApprovalRequestByIdempotencyKeyParams{
		WorkspaceID:    workspaceID,
		IdempotencyKey: cmd.IdempotencyKey,
	})

	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return SubmitResult{}, fmt.Errorf("commit idempotent approval request transaction: %w", err)
		}

		return SubmitResult{Request: existing, Existing: true}, nil
	}
	if err != nil && err != pgx.ErrNoRows {
		return SubmitResult{}, fmt.Errorf("load approval request by idempotency key: %w", err)
	}

	workspace, err := queries.GetWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("load workspace: %w", err)
	}

	actionJSON, actionHash, err := canonicalizeAction(cmd.Action)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("canonicalize action: %w", err)
	}

	contextJSON, err := marshalJSONObject(map[string]any{
		"reason":     cmd.Context.Reason,
		"evidence":   rawMessageOrNil(cmd.Context.Evidence),
		"confidence": cmd.Context.Confidence,
		"reversible": cmd.Context.Reversible,
		"deadline":   cmd.Context.Deadline,
		"metadata":   cmd.Context.Metadata,
	})
	if err != nil {
		return SubmitResult{}, fmt.Errorf("marshal request context: %w", err)
	}

	affectedSystemsJSON, err := json.Marshal(cmd.Context.AffectedSystems)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("marshal affected systems: %w", err)
	}

	requestMetadataJSON, err := marshalJSONObject(map[string]any{
		"submitted_via": "api",
	})
	if err != nil {
		return SubmitResult{}, fmt.Errorf("marshal request metadata: %w", err)
	}

	now := s.clock().UTC()

	evaluation, err := evaluatePolicies(ctx, queries, workspaceID, workspace.DefaultPolicyEffect, cmd, now)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("evaluate policies: %w", err)
	}

	request, err := queries.CreateApprovalRequest(ctx, generated.CreateApprovalRequestParams{
		WorkspaceID:            workspaceID,
		IdempotencyKey:         cmd.IdempotencyKey,
		ActionType:             cmd.Action.Type,
		Title:                  cmd.Action.Title,
		Description:            pgxutil.Text(cmd.Action.Description),
		OriginalAction:         string(actionJSON),
		OriginalActionHash:     actionHash,
		SourcePlatform:         cmd.Source.Platform,
		SourceWorkflowID:       cmd.Source.WorkflowID,
		SourceExecutionID:      cmd.Source.ExecutionID,
		Context:                string(contextJSON),
		AffectedSystems:        string(affectedSystemsJSON),
		Metadata:               string(requestMetadataJSON),
		MatchedPolicyID:        evaluation.PolicyID,
		MatchedPolicyVersionID: evaluation.PolicyVersionID,
		MatchedPolicySnapshot:  jsonbStringOrNull(evaluation.PolicySnapshot),
		AssignedUserID:         evaluation.AssignedUserID,
		AssignedGroupID:        evaluation.AssignedGroupID,
		Status:                 evaluation.Status,
		DecisionRequired:       evaluation.DecisionRequired,
		ExpiresAt:              evaluation.ExpiresAt,
		ResolvedAt:             evaluation.ResolvedAt,
	})
	if err != nil {
		return SubmitResult{}, fmt.Errorf("create approval request: %w", err)
	}

	continuationTarget, err := queries.CreateContinuationTarget(ctx, generated.CreateContinuationTargetParams{
		WorkspaceID:            workspaceID,
		ApprovalRequestID:      request.ID,
		Strategy:               generated.ContinuationStrategy(cmd.Continuation.Strategy),
		Platform:               cmd.Source.Platform,
		Destination:            pgxutil.Text(cmd.Continuation.URL),
		EncryptedConfiguration: `{}`,
	})
	if err != nil {
		return SubmitResult{}, fmt.Errorf("create continuation target: %w", err)
	}

	actorID, err := optionalUUID(cmd.AuthenticatedBy.ActorID)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("parse actor id: %w", err)
	}

	auditMetadata, err := marshalJSONObject(map[string]any{
		"status":          request.Status,
		"source_platform": request.SourcePlatform,
		"idempotency_key": request.IdempotencyKey,
	})
	if err != nil {
		return SubmitResult{}, fmt.Errorf("marshal audit metadata: %w", err)
	}

	if _, err := queries.CreateAuditEvent(ctx, generated.CreateAuditEventParams{
		WorkspaceID:       workspaceID,
		ApprovalRequestID: request.ID,
		DecisionID:        pgtype.UUID{},
		ActorType:         cmd.AuthenticatedBy.ActorType,
		ActorID:           actorID,
		EventType:         "approval_request.received",
		Metadata:          string(auditMetadata),
	}); err != nil {
		return SubmitResult{}, fmt.Errorf("create audit event: %w", err)
	}

	if err := createAutomaticPolicyDecision(ctx, queries, request, continuationTarget, now); err != nil {
		return SubmitResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return SubmitResult{}, fmt.Errorf("commit approval request transaction: %w", err)
	}

	return SubmitResult{Request: request}, nil
}

func createAutomaticPolicyDecision(ctx context.Context, queries *generated.Queries, request generated.ApprovalRequest, continuationTarget generated.ContinuationTarget, now time.Time) error {
	decision, eventType, comment, approvedAction, approvedActionHash, ok := automaticPolicyDecision(request)
	if !ok {
		return nil
	}

	createdDecision, err := queries.CreateApprovalDecision(ctx, generated.CreateApprovalDecisionParams{
		WorkspaceID:        request.WorkspaceID,
		ApprovalRequestID:  request.ID,
		Decision:           decision,
		OriginalActionHash: request.OriginalActionHash,
		ApprovedAction:     approvedAction,
		ApprovedActionHash: approvedActionHash,
		ChangedFields:      `[]`,
		Comment:            pgxutil.Text(comment),
		DecidedBy:          pgtype.UUID{},
		IssuedAt:           pgxutil.Timestamptz(now),
		ExpiresAt:          pgtype.Timestamptz{},
	})
	if err != nil {
		return fmt.Errorf("create automatic policy decision: %w", err)
	}

	createdDelivery, err := queries.CreateDecisionDelivery(ctx, generated.CreateDecisionDeliveryParams{
		WorkspaceID:          request.WorkspaceID,
		DecisionID:           createdDecision.ID,
		ContinuationTargetID: continuationTarget.ID,
		Status:               generated.DeliveryStatusPending,
		NextAttemptAt:        pgxutil.Timestamptz(now),
	})
	if err != nil {
		return fmt.Errorf("create automatic policy delivery: %w", err)
	}

	decisionAuditMetadata, err := marshalJSONObject(map[string]any{
		"decision":  decision,
		"automatic": true,
	})
	if err != nil {
		return fmt.Errorf("marshal automatic policy audit metadata: %w", err)
	}

	if _, err := queries.CreateAuditEvent(ctx, generated.CreateAuditEventParams{
		WorkspaceID:       request.WorkspaceID,
		ApprovalRequestID: request.ID,
		DecisionID:        createdDecision.ID,
		ActorType:         "system",
		ActorID:           pgtype.UUID{},
		EventType:         eventType,
		Metadata:          string(decisionAuditMetadata),
	}); err != nil {
		return fmt.Errorf("create automatic policy audit event: %w", err)
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
		return fmt.Errorf("marshal automatic policy delivery audit metadata: %w", err)
	}

	if _, err := queries.CreateAuditEvent(ctx, generated.CreateAuditEventParams{
		WorkspaceID:       request.WorkspaceID,
		ApprovalRequestID: request.ID,
		DecisionID:        createdDecision.ID,
		ActorType:         "system",
		ActorID:           pgtype.UUID{},
		EventType:         "decision.delivery_scheduled",
		Metadata:          string(deliveryAuditMetadata),
	}); err != nil {
		return fmt.Errorf("create automatic policy delivery audit event: %w", err)
	}

	return nil
}

func automaticPolicyDecision(request generated.ApprovalRequest) (generated.DecisionType, string, string, string, pgtype.Text, bool) {
	switch request.Status {
	case generated.ApprovalStatusAllowed:
		return generated.DecisionTypeAllowed,
			"approval_request.allowed",
			"Allowed automatically by policy.",
			string(request.OriginalAction),
			pgxutil.Text(request.OriginalActionHash),
			true
	case generated.ApprovalStatusBlocked:
		return generated.DecisionTypeBlocked,
			"approval_request.blocked",
			"Blocked automatically by policy.",
			`null`,
			pgtype.Text{},
			true
	default:
		return "", "", "", "", pgtype.Text{}, false
	}
}

func validateSubmitCommand(cmd SubmitApprovalRequestCommand) error {
	switch {
	case cmd.WorkspaceID == "":
		return ErrInvalidRequest
	case cmd.IdempotencyKey == "":
		return ErrInvalidRequest
	case cmd.Action.Type == "":
		return ErrInvalidRequest
	case cmd.Action.Title == "":
		return ErrInvalidRequest
	case cmd.Source.Platform == "":
		return ErrInvalidRequest
	case cmd.Source.WorkflowID == "":
		return ErrInvalidRequest
	case cmd.Source.ExecutionID == "":
		return ErrInvalidRequest
	case cmd.Continuation.Strategy == "":
		return ErrInvalidRequest
	default:
		return nil
	}
}

type policyEvaluation struct {
	Status           generated.ApprovalStatus
	DecisionRequired bool
	PolicyID         pgtype.UUID
	PolicyVersionID  pgtype.UUID
	PolicySnapshot   []byte
	AssignedUserID   pgtype.UUID
	AssignedGroupID  pgtype.UUID
	ExpiresAt        pgtype.Timestamptz
	ResolvedAt       pgtype.Timestamptz
}

func evaluatePolicies(
	ctx context.Context,
	queries *generated.Queries,
	workspaceID pgtype.UUID,
	defaultEffect generated.PolicyEffect,
	cmd SubmitApprovalRequestCommand,
	now time.Time,
) (policyEvaluation, error) {
	policies, err := queries.ListActivePolicyVersionsForWorkspace(ctx, workspaceID)
	if err != nil {
		return policyEvaluation{}, err
	}

	for _, policy := range policies {
		match, assignedUserID, assignedGroupID, expiresAt, err := matchesPolicy(policy, cmd, now)
		if err != nil {
			return policyEvaluation{}, err
		}
		if !match {
			continue
		}

		snapshot, err := marshalJSONObject(map[string]any{
			"policy_id":         pgxutil.UUIDString(policy.PolicyID),
			"policy_version_id": pgxutil.UUIDString(policy.PolicyVersionID),
			"name":              policy.Name,
			"priority":          policy.Priority,
			"version_number":    policy.VersionNumber,
			"effect":            policy.Effect,
			"conditions":        rawMessageOrNil(policy.Conditions),
			"approval_settings": rawMessageOrNil(policy.ApprovalSettings),
		})
		if err != nil {
			return policyEvaluation{}, err
		}

		return effectToEvaluation(policy.Effect, policy.PolicyID, policy.PolicyVersionID, snapshot, assignedUserID, assignedGroupID, expiresAt, now), nil
	}

	return effectToEvaluation(defaultEffect, pgtype.UUID{}, pgtype.UUID{}, nil, pgtype.UUID{}, pgtype.UUID{}, pgtype.Timestamptz{}, now), nil
}

func effectToEvaluation(effect generated.PolicyEffect, policyID pgtype.UUID, policyVersionID pgtype.UUID, snapshot []byte, assignedUserID pgtype.UUID, assignedGroupID pgtype.UUID, expiresAt pgtype.Timestamptz, now time.Time) policyEvaluation {
	switch effect {
	case generated.PolicyEffectAllow:
		return policyEvaluation{
			Status:           generated.ApprovalStatusAllowed,
			DecisionRequired: false,
			PolicyID:         policyID,
			PolicyVersionID:  policyVersionID,
			PolicySnapshot:   snapshot,
			ResolvedAt:       pgxutil.Timestamptz(now),
		}
	case generated.PolicyEffectBlock:
		return policyEvaluation{
			Status:           generated.ApprovalStatusBlocked,
			DecisionRequired: false,
			PolicyID:         policyID,
			PolicyVersionID:  policyVersionID,
			PolicySnapshot:   snapshot,
			ResolvedAt:       pgxutil.Timestamptz(now),
		}
	default:
		return policyEvaluation{
			Status:           generated.ApprovalStatusPending,
			DecisionRequired: true,
			PolicyID:         policyID,
			PolicyVersionID:  policyVersionID,
			PolicySnapshot:   snapshot,
			AssignedUserID:   assignedUserID,
			AssignedGroupID:  assignedGroupID,
			ExpiresAt:        expiresAt,
		}
	}
}

type policyCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type approvalSettings struct {
	ApproverUserID  string `json:"approver_user_id"`
	ApproverGroupID string `json:"approver_group_id"`
	DeadlineSeconds int64  `json:"deadline_seconds"`
}

func matchesPolicy(policy generated.ListActivePolicyVersionsForWorkspaceRow, cmd SubmitApprovalRequestCommand, now time.Time) (bool, pgtype.UUID, pgtype.UUID, pgtype.Timestamptz, error) {
	var conditions []policyCondition
	if len(policy.Conditions) > 0 {
		if err := json.Unmarshal(policy.Conditions, &conditions); err != nil {
			return false, pgtype.UUID{}, pgtype.UUID{}, pgtype.Timestamptz{}, fmt.Errorf("unmarshal policy conditions: %w", err)
		}
	}
	for _, condition := range conditions {
		if !evaluateCondition(condition, cmd) {
			return false, pgtype.UUID{}, pgtype.UUID{}, pgtype.Timestamptz{}, nil
		}
	}

	var settings approvalSettings
	if len(policy.ApprovalSettings) > 0 {
		if err := json.Unmarshal(policy.ApprovalSettings, &settings); err != nil {
			return false, pgtype.UUID{}, pgtype.UUID{}, pgtype.Timestamptz{}, fmt.Errorf("unmarshal approval settings: %w", err)
		}
	}

	assignedUserID, err := optionalUUID(settings.ApproverUserID)
	if err != nil {
		return false, pgtype.UUID{}, pgtype.UUID{}, pgtype.Timestamptz{}, fmt.Errorf("parse approver user id: %w", err)
	}

	assignedGroupID, err := optionalUUID(settings.ApproverGroupID)
	if err != nil {
		return false, pgtype.UUID{}, pgtype.UUID{}, pgtype.Timestamptz{}, fmt.Errorf("parse approver group id: %w", err)
	}

	var expiresAt pgtype.Timestamptz
	if settings.DeadlineSeconds > 0 {
		expiresAt = pgxutil.Timestamptz(now.Add(time.Duration(settings.DeadlineSeconds) * time.Second))
	}

	return true, assignedUserID, assignedGroupID, expiresAt, nil
}

func evaluateCondition(condition policyCondition, cmd SubmitApprovalRequestCommand) bool {
	switch {
	case condition.Field == "action.type" && condition.Operator == "equals":
		value, _ := condition.Value.(string)
		return cmd.Action.Type == value
	case condition.Field == "source.platform" && condition.Operator == "equals":
		value, _ := condition.Value.(string)
		return cmd.Source.Platform == value
	case condition.Field == "context.reversible" && condition.Operator == "equals":
		value, ok := condition.Value.(bool)
		if !ok || cmd.Context.Reversible == nil {
			return false
		}
		return *cmd.Context.Reversible == value
	default:
		return false
	}
}

func canonicalizeAction(action ProposedAction) ([]byte, string, error) {
	payload, err := marshalJSONObject(map[string]any{
		"type":        action.Type,
		"title":       action.Title,
		"description": action.Description,
		"parameters":  rawMessageOrNil(action.Parameters),
	})
	if err != nil {
		return nil, "", err
	}

	sum := sha256.Sum256(payload)
	return payload, hex.EncodeToString(sum[:]), nil
}

func marshalJSONObject(value map[string]any) ([]byte, error) {
	var normalized interface{}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()

	if err := decoder.Decode(&normalized); err != nil {
		return nil, err
	}

	return json.Marshal(normalized)
}

func rawMessageOrNil(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}

	return json.RawMessage(raw)
}

func jsonbStringOrNull(raw []byte) string {
	if len(raw) == 0 {
		return `null`
	}

	return string(raw)
}

func optionalUUID(value string) (pgtype.UUID, error) {
	if value == "" {
		return pgtype.UUID{}, nil
	}

	parsed, err := uuid.Parse(value)
	if err != nil {
		return pgtype.UUID{}, err
	}

	return pgxutil.UUID(parsed), nil
}
