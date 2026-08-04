package apikeys

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"greenpost/db/generated"
	"greenpost/internal/identity"
	"greenpost/internal/platform/pgxutil"

	"github.com/jackc/pgx/v5"
)

var ErrInvalidAPIKeyRequest = errors.New("invalid api key request")

type CreateAPIKeyCommand struct {
	WorkspaceID string
	UserID      string
	Name        string
	Scopes      []string
}

type CreateAPIKeyResult struct {
	APIKey generated.ApiKey
	RawKey string
}

func (s *Service) CreateAPIKey(ctx context.Context, cmd CreateAPIKeyCommand) (CreateAPIKeyResult, error) {
	if strings.TrimSpace(cmd.WorkspaceID) == "" || strings.TrimSpace(cmd.UserID) == "" || strings.TrimSpace(cmd.Name) == "" {
		return CreateAPIKeyResult{}, ErrInvalidAPIKeyRequest
	}

	workspaceID, err := pgxutil.UUIDText(cmd.WorkspaceID)
	if err != nil {
		return CreateAPIKeyResult{}, identity.ErrInvalidWorkspaceID
	}

	userID, err := pgxutil.UUIDText(cmd.UserID)
	if err != nil {
		return CreateAPIKeyResult{}, identity.ErrInvalidUserID
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return CreateAPIKeyResult{}, fmt.Errorf("begin api key transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := generated.New(tx)

	member, err := queries.GetWorkspaceMember(ctx, generated.GetWorkspaceMemberParams{
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CreateAPIKeyResult{}, identity.ErrForbidden
		}
		return CreateAPIKeyResult{}, fmt.Errorf("load workspace member: %w", err)
	}

	if member.Role != generated.WorkspaceRoleOwner && member.Role != generated.WorkspaceRoleAdministrator {
		return CreateAPIKeyResult{}, identity.ErrForbidden
	}

	rawKey, prefix, err := generateRawAPIKey()
	if err != nil {
		return CreateAPIKeyResult{}, fmt.Errorf("generate api key: %w", err)
	}

	created, err := queries.CreateAPIKey(ctx, generated.CreateAPIKeyParams{
		WorkspaceID: workspaceID,
		Name:        strings.TrimSpace(cmd.Name),
		KeyPrefix:   prefix,
		KeyHash:     HashAPIKey(rawKey),
		Scopes:      normalizeScopes(cmd.Scopes),
		CreatedBy:   userID,
	})
	if err != nil {
		return CreateAPIKeyResult{}, fmt.Errorf("create api key: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return CreateAPIKeyResult{}, fmt.Errorf("commit api key transaction: %w", err)
	}

	return CreateAPIKeyResult{
		APIKey: created,
		RawKey: rawKey,
	}, nil
}

func generateRawAPIKey() (rawKey string, prefix string, err error) {
	prefixBytes := make([]byte, 6)
	secretBytes := make([]byte, 24)

	if _, err := rand.Read(prefixBytes); err != nil {
		return "", "", err
	}

	if _, err := rand.Read(secretBytes); err != nil {
		return "", "", err
	}

	prefix = "hgk_" + hex.EncodeToString(prefixBytes)
	secret := hex.EncodeToString(secretBytes)

	return prefix + "." + secret, prefix, nil
}

func normalizeScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return []string{"approval_requests:create"}
	}

	seen := make(map[string]struct{}, len(scopes))
	normalized := make([]string, 0, len(scopes))

	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if _, exists := seen[scope]; exists {
			continue
		}
		seen[scope] = struct{}{}
		normalized = append(normalized, scope)
	}

	return normalized
}
