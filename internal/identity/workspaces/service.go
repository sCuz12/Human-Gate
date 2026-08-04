package workspaces

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"greenpost/db/generated"
	"greenpost/internal/identity"
	"greenpost/internal/platform/pgxutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var nonSlugPattern = regexp.MustCompile(`[^a-z0-9]+`)

type Service struct {
	db *pgxpool.Pool
}

type CreateWorkspaceCommand struct {
	UserID string
	Name   string
}

type WorkspaceSummary struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Slug                string `json:"slug"`
	DefaultPolicyEffect string `json:"default_policy_effect"`
	Role                string `json:"role"`
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) CreateWorkspace(ctx context.Context, cmd CreateWorkspaceCommand) (WorkspaceSummary, error) {
	if strings.TrimSpace(cmd.UserID) == "" || strings.TrimSpace(cmd.Name) == "" {
		return WorkspaceSummary{}, ErrInvalidWorkspaceRequest
	}

	userID, err := pgxutil.UUIDText(cmd.UserID)
	if err != nil {
		return WorkspaceSummary{}, identity.ErrInvalidUserID
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WorkspaceSummary{}, fmt.Errorf("begin workspace transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := generated.New(tx)

	baseSlug := slugify(cmd.Name)
	if baseSlug == "" {
		return WorkspaceSummary{}, ErrInvalidWorkspaceRequest
	}

	slug, err := s.allocateSlug(ctx, queries, baseSlug)
	if err != nil {
		return WorkspaceSummary{}, fmt.Errorf("allocate workspace slug: %w", err)
	}

	workspace, err := queries.CreateWorkspace(ctx, generated.CreateWorkspaceParams{
		Name:                strings.TrimSpace(cmd.Name),
		Slug:                slug,
		DefaultPolicyEffect: generated.PolicyEffectRequireApproval,
		CreatedBy:           userID,
	})
	if err != nil {
		return WorkspaceSummary{}, fmt.Errorf("create workspace: %w", err)
	}

	member, err := queries.CreateWorkspaceMember(ctx, generated.CreateWorkspaceMemberParams{
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Role:        generated.WorkspaceRoleOwner,
	})
	if err != nil {
		return WorkspaceSummary{}, fmt.Errorf("create workspace membership: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return WorkspaceSummary{}, fmt.Errorf("commit workspace transaction: %w", err)
	}

	return toWorkspaceSummary(workspace, member.Role), nil
}

func (s *Service) ListWorkspaces(ctx context.Context, userID string) ([]WorkspaceSummary, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalidWorkspaceRequest
	}

	parsedUserID, err := pgxutil.UUIDText(userID)
	if err != nil {
		return nil, identity.ErrInvalidUserID
	}

	queries := generated.New(s.db)
	rows, err := queries.ListWorkspacesByUserID(ctx, parsedUserID)
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}

	workspaces := make([]WorkspaceSummary, 0, len(rows))
	for _, row := range rows {
		workspaces = append(workspaces, WorkspaceSummary{
			ID:                  pgxutil.UUIDString(row.ID),
			Name:                row.Name,
			Slug:                row.Slug,
			DefaultPolicyEffect: string(row.DefaultPolicyEffect),
			Role:                string(row.Role),
		})
	}

	return workspaces, nil
}

func (s *Service) allocateSlug(ctx context.Context, queries *generated.Queries, baseSlug string) (string, error) {
	slug := baseSlug
	for suffix := 0; suffix < 100; suffix++ {
		if suffix > 0 {
			slug = fmt.Sprintf("%s-%d", baseSlug, suffix+1)
		}

		_, err := queries.GetWorkspaceBySlug(ctx, slug)
		if err == nil {
			continue
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return slug, nil
		}
		return "", err
	}

	return "", fmt.Errorf("unable to allocate unique slug for base %q", baseSlug)
}

func slugify(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = nonSlugPattern.ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	return normalized
}

func toWorkspaceSummary(workspace generated.Workspace, role generated.WorkspaceRole) WorkspaceSummary {
	return WorkspaceSummary{
		ID:                  pgxutil.UUIDString(workspace.ID),
		Name:                workspace.Name,
		Slug:                workspace.Slug,
		DefaultPolicyEffect: string(workspace.DefaultPolicyEffect),
		Role:                string(role),
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
