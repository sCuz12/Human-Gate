package policy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"decree/db/generated"
	"decree/internal/identity"
	"decree/internal/platform/pgxutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

type Condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

type CreatePolicyCommand struct {
	WorkspaceID     string
	UserID          string
	Name            string
	Description     string
	Priority        int32
	IsActive        bool
	Conditions      []Condition
	Effect          string
	DeadlineSeconds int64
}

type UpdatePolicyCommand struct {
	WorkspaceID     string
	PolicyID        string
	UserID          string
	Name            string
	Description     string
	Priority        int32
	IsActive        bool
	Conditions      []Condition
	Effect          string
	DeadlineSeconds int64
}

type DeletePolicyCommand struct {
	WorkspaceID string
	PolicyID    string
	UserID      string
}

type PolicySummary struct {
	ID               string          `json:"id"`
	WorkspaceID      string          `json:"workspace_id"`
	Name             string          `json:"name"`
	Description      string          `json:"description,omitempty"`
	Priority         int32           `json:"priority"`
	IsActive         bool            `json:"is_active"`
	VersionID        string          `json:"version_id"`
	VersionNumber    int32           `json:"version_number"`
	Conditions       json.RawMessage `json:"conditions"`
	Effect           string          `json:"effect"`
	ApprovalSettings json.RawMessage `json:"approval_settings"`
	DeadlineSeconds  int64           `json:"deadline_seconds,omitempty"`
	CreatedAt        string          `json:"created_at"`
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) CreatePolicy(ctx context.Context, cmd CreatePolicyCommand) (PolicySummary, error) {
	if strings.TrimSpace(cmd.WorkspaceID) == "" || strings.TrimSpace(cmd.UserID) == "" || strings.TrimSpace(cmd.Name) == "" {
		return PolicySummary{}, ErrInvalidRequest
	}

	workspaceID, err := pgxutil.UUIDText(cmd.WorkspaceID)
	if err != nil {
		return PolicySummary{}, identity.ErrInvalidWorkspaceID
	}

	userID, err := pgxutil.UUIDText(cmd.UserID)
	if err != nil {
		return PolicySummary{}, identity.ErrInvalidUserID
	}

	effect, err := parseEffect(cmd.Effect)
	if err != nil {
		return PolicySummary{}, err
	}

	if cmd.Priority <= 0 {
		cmd.Priority = 100
	}

	conditions, err := normalizeConditions(cmd.Conditions)
	if err != nil {
		return PolicySummary{}, err
	}

	approvalSettings, err := normalizeApprovalSettings(effect, cmd.DeadlineSeconds)
	if err != nil {
		return PolicySummary{}, err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return PolicySummary{}, fmt.Errorf("begin policy transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := generated.New(tx)
	if err := requirePolicyManager(ctx, queries, workspaceID, userID); err != nil {
		return PolicySummary{}, err
	}

	createdPolicy, err := queries.CreatePolicy(ctx, generated.CreatePolicyParams{
		WorkspaceID: workspaceID,
		Name:        strings.TrimSpace(cmd.Name),
		Description: pgxutil.Text(cmd.Description),
		Priority:    cmd.Priority,
		IsActive:    cmd.IsActive,
		CreatedBy:   userID,
	})
	if err != nil {
		return PolicySummary{}, fmt.Errorf("create policy: %w", err)
	}

	version, err := queries.CreatePolicyVersion(ctx, generated.CreatePolicyVersionParams{
		WorkspaceID:      workspaceID,
		PolicyID:         createdPolicy.ID,
		VersionNumber:    1,
		Conditions:       string(conditions),
		Effect:           effect,
		ApprovalSettings: string(approvalSettings),
		CreatedBy:        userID,
	})
	if err != nil {
		return PolicySummary{}, fmt.Errorf("create policy version: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return PolicySummary{}, fmt.Errorf("commit policy transaction: %w", err)
	}

	return policySummaryFromPolicy(createdPolicy, version), nil
}

func (s *Service) ListPolicies(ctx context.Context, workspaceIDValue string, userIDValue string) ([]PolicySummary, error) {
	if strings.TrimSpace(workspaceIDValue) == "" || strings.TrimSpace(userIDValue) == "" {
		return nil, ErrInvalidRequest
	}

	workspaceID, err := pgxutil.UUIDText(workspaceIDValue)
	if err != nil {
		return nil, identity.ErrInvalidWorkspaceID
	}

	userID, err := pgxutil.UUIDText(userIDValue)
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

	rows, err := queries.ListPolicySummariesForWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}

	policies := make([]PolicySummary, 0, len(rows))
	for _, row := range rows {
		policies = append(policies, policySummaryFromRow(row))
	}

	return policies, nil
}

func (s *Service) UpdatePolicy(ctx context.Context, cmd UpdatePolicyCommand) (PolicySummary, error) {
	if strings.TrimSpace(cmd.WorkspaceID) == "" || strings.TrimSpace(cmd.PolicyID) == "" || strings.TrimSpace(cmd.UserID) == "" || strings.TrimSpace(cmd.Name) == "" {
		return PolicySummary{}, ErrInvalidRequest
	}

	workspaceID, policyID, userID, err := parsePolicyMutationIDs(cmd.WorkspaceID, cmd.PolicyID, cmd.UserID)
	if err != nil {
		return PolicySummary{}, err
	}

	effect, err := parseEffect(cmd.Effect)
	if err != nil {
		return PolicySummary{}, err
	}

	if cmd.Priority <= 0 {
		cmd.Priority = 100
	}

	conditions, err := normalizeConditions(cmd.Conditions)
	if err != nil {
		return PolicySummary{}, err
	}

	approvalSettings, err := normalizeApprovalSettings(effect, cmd.DeadlineSeconds)
	if err != nil {
		return PolicySummary{}, err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return PolicySummary{}, fmt.Errorf("begin policy update transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := generated.New(tx)
	if err := requirePolicyManager(ctx, queries, workspaceID, userID); err != nil {
		return PolicySummary{}, err
	}

	latest, err := queries.GetLatestPolicyVersionForUpdate(ctx, generated.GetLatestPolicyVersionForUpdateParams{
		WorkspaceID: workspaceID,
		PolicyID:    policyID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PolicySummary{}, ErrNotFound
		}
		return PolicySummary{}, fmt.Errorf("load latest policy version: %w", err)
	}

	updatedPolicy, err := queries.UpdatePolicy(ctx, generated.UpdatePolicyParams{
		Name:        strings.TrimSpace(cmd.Name),
		Description: pgxutil.Text(cmd.Description),
		Priority:    cmd.Priority,
		IsActive:    cmd.IsActive,
		UpdatedAt:   pgxutil.Timestamptz(time.Now().UTC()),
		WorkspaceID: workspaceID,
		ID:          policyID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PolicySummary{}, ErrNotFound
		}
		return PolicySummary{}, fmt.Errorf("update policy: %w", err)
	}

	version, err := queries.CreatePolicyVersion(ctx, generated.CreatePolicyVersionParams{
		WorkspaceID:      workspaceID,
		PolicyID:         policyID,
		VersionNumber:    latest.VersionNumber + 1,
		Conditions:       string(conditions),
		Effect:           effect,
		ApprovalSettings: string(approvalSettings),
		CreatedBy:        userID,
	})
	if err != nil {
		return PolicySummary{}, fmt.Errorf("create updated policy version: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return PolicySummary{}, fmt.Errorf("commit policy update transaction: %w", err)
	}

	return policySummaryFromPolicy(updatedPolicy, version), nil
}

func (s *Service) DeletePolicy(ctx context.Context, cmd DeletePolicyCommand) error {
	if strings.TrimSpace(cmd.WorkspaceID) == "" || strings.TrimSpace(cmd.PolicyID) == "" || strings.TrimSpace(cmd.UserID) == "" {
		return ErrInvalidRequest
	}

	workspaceID, policyID, userID, err := parsePolicyMutationIDs(cmd.WorkspaceID, cmd.PolicyID, cmd.UserID)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin policy delete transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := generated.New(tx)
	if err := requirePolicyManager(ctx, queries, workspaceID, userID); err != nil {
		return err
	}

	now := time.Now().UTC()
	if _, err := queries.SoftDeletePolicy(ctx, generated.SoftDeletePolicyParams{
		DeletedAt:   pgxutil.Timestamptz(now),
		WorkspaceID: workspaceID,
		ID:          policyID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("soft delete policy: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit policy delete transaction: %w", err)
	}

	return nil
}

func requirePolicyManager(ctx context.Context, queries *generated.Queries, workspaceID pgtype.UUID, userID pgtype.UUID) error {
	member, err := queries.GetWorkspaceMember(ctx, generated.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrForbidden
		}
		return fmt.Errorf("load workspace member: %w", err)
	}

	switch member.Role {
	case generated.WorkspaceRoleOwner, generated.WorkspaceRoleAdministrator:
		return nil
	default:
		return ErrForbidden
	}
}

func parsePolicyMutationIDs(workspaceIDValue string, policyIDValue string, userIDValue string) (pgtype.UUID, pgtype.UUID, pgtype.UUID, error) {
	workspaceID, err := pgxutil.UUIDText(workspaceIDValue)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, identity.ErrInvalidWorkspaceID
	}

	policyID, err := pgxutil.UUIDText(policyIDValue)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, ErrInvalidRequest
	}

	userID, err := pgxutil.UUIDText(userIDValue)
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, identity.ErrInvalidUserID
	}

	return workspaceID, policyID, userID, nil
}

func parseEffect(effect string) (generated.PolicyEffect, error) {
	switch strings.TrimSpace(effect) {
	case "allow":
		return generated.PolicyEffectAllow, nil
	case "require_approval":
		return generated.PolicyEffectRequireApproval, nil
	case "block":
		return generated.PolicyEffectBlock, nil
	default:
		return "", ErrInvalidRequest
	}
}

func normalizeConditions(conditions []Condition) ([]byte, error) {
	normalized := make([]Condition, 0, len(conditions))
	for _, condition := range conditions {
		field := strings.TrimSpace(condition.Field)
		operator := strings.TrimSpace(condition.Operator)
		if field == "" && operator == "" && condition.Value == nil {
			continue
		}

		if operator != "equals" {
			return nil, ErrInvalidRequest
		}

		switch field {
		case "action.type", "source.platform":
			value, ok := condition.Value.(string)
			if !ok || strings.TrimSpace(value) == "" {
				return nil, ErrInvalidRequest
			}
			normalized = append(normalized, Condition{Field: field, Operator: operator, Value: strings.TrimSpace(value)})
		case "context.reversible":
			value, ok := condition.Value.(bool)
			if !ok {
				return nil, ErrInvalidRequest
			}
			normalized = append(normalized, Condition{Field: field, Operator: operator, Value: value})
		default:
			return nil, ErrInvalidRequest
		}
	}

	raw, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal policy conditions: %w", err)
	}
	return raw, nil
}

func normalizeApprovalSettings(effect generated.PolicyEffect, deadlineSeconds int64) ([]byte, error) {
	if deadlineSeconds < 0 {
		return nil, ErrInvalidRequest
	}
	if effect != generated.PolicyEffectRequireApproval {
		deadlineSeconds = 0
	}

	raw, err := json.Marshal(map[string]any{
		"deadline_seconds": deadlineSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal approval settings: %w", err)
	}
	return raw, nil
}

func policySummaryFromPolicy(policy generated.Policy, version generated.PolicyVersion) PolicySummary {
	return PolicySummary{
		ID:               pgxutil.UUIDString(policy.ID),
		WorkspaceID:      pgxutil.UUIDString(policy.WorkspaceID),
		Name:             policy.Name,
		Description:      policy.Description.String,
		Priority:         policy.Priority,
		IsActive:         policy.IsActive,
		VersionID:        pgxutil.UUIDString(version.ID),
		VersionNumber:    version.VersionNumber,
		Conditions:       json.RawMessage(version.Conditions),
		Effect:           string(version.Effect),
		ApprovalSettings: json.RawMessage(version.ApprovalSettings),
		DeadlineSeconds:  deadlineSeconds(version.ApprovalSettings),
		CreatedAt:        policy.CreatedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func policySummaryFromRow(row generated.ListPolicySummariesForWorkspaceRow) PolicySummary {
	return PolicySummary{
		ID:               pgxutil.UUIDString(row.PolicyID),
		WorkspaceID:      pgxutil.UUIDString(row.WorkspaceID),
		Name:             row.Name,
		Description:      row.Description.String,
		Priority:         row.Priority,
		IsActive:         row.IsActive,
		VersionID:        pgxutil.UUIDString(row.PolicyVersionID),
		VersionNumber:    row.VersionNumber,
		Conditions:       json.RawMessage(row.Conditions),
		Effect:           string(row.Effect),
		ApprovalSettings: json.RawMessage(row.ApprovalSettings),
		DeadlineSeconds:  deadlineSeconds(row.ApprovalSettings),
		CreatedAt:        row.CreatedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func deadlineSeconds(raw []byte) int64 {
	var settings struct {
		DeadlineSeconds int64 `json:"deadline_seconds"`
	}
	if len(raw) == 0 {
		return 0
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return 0
	}
	return settings.DeadlineSeconds
}
