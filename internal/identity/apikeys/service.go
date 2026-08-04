package apikeys

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"greenpost/db/generated"
	"greenpost/internal/platform/pgxutil"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Clock func() time.Time

type Service struct {
	queries *generated.Queries
	db      *pgxpool.Pool
	clock   Clock
}

type AuthenticatedKey struct {
	APIKey    generated.ApiKey
	Workspace generated.Workspace
}

func NewService(db *pgxpool.Pool, clock Clock) *Service {
	if clock == nil {
		clock = time.Now
	}

	return &Service{
		queries: generated.New(db),
		db:      db,
		clock:   clock,
	}
}

func (s *Service) Authenticate(ctx context.Context, rawKey string, requiredScope string) (AuthenticatedKey, error) {
	rawKey = strings.TrimSpace(rawKey)
	if rawKey == "" {
		return AuthenticatedKey{}, ErrMissingAPIKey
	}

	prefix, ok := extractPrefix(rawKey)
	if !ok {
		return AuthenticatedKey{}, ErrInvalidAPIKey
	}

	candidates, err := s.queries.GetActiveAPIKeyByPrefixGlobal(ctx, prefix)
	if err != nil {
		return AuthenticatedKey{}, fmt.Errorf("load api key candidates: %w", err)
	}

	hashed := hashAPIKey(rawKey)

	for _, candidate := range candidates {
		if subtle.ConstantTimeCompare([]byte(candidate.KeyHash), []byte(hashed)) != 1 {
			continue
		}

		if requiredScope != "" && !hasScope(candidate.Scopes, requiredScope) {
			return AuthenticatedKey{}, ErrAPIKeyUnauthorized
		}

		workspace, err := s.queries.GetWorkspaceByID(ctx, candidate.WorkspaceID)
		if err != nil {
			return AuthenticatedKey{}, fmt.Errorf("load workspace for api key: %w", err)
		}

		if err := s.queries.UpdateAPIKeyLastUsedAt(ctx, generated.UpdateAPIKeyLastUsedAtParams{
			WorkspaceID: candidate.WorkspaceID,
			ID:          candidate.ID,
			LastUsedAt:  pgxutil.Timestamptz(s.clock()),
		}); err != nil {
			return AuthenticatedKey{}, fmt.Errorf("update api key last used: %w", err)
		}

		return AuthenticatedKey{
			APIKey:    candidate,
			Workspace: workspace,
		}, nil
	}

	return AuthenticatedKey{}, ErrInvalidAPIKey
}

func HashAPIKey(rawKey string) string {
	return hashAPIKey(rawKey)
}

func hashAPIKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}

func extractPrefix(rawKey string) (string, bool) {
	prefix, _, ok := strings.Cut(rawKey, ".")
	if !ok || prefix == "" {
		return "", false
	}

	return prefix, true
}

func hasScope(scopes []string, requiredScope string) bool {
	for _, scope := range scopes {
		if scope == requiredScope {
			return true
		}
	}

	return false
}
