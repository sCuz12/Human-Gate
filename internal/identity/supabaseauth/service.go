package supabaseauth

import (
	"context"
	"fmt"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	keyfunc jwt.Keyfunc
	issuer  string
}

type UserClaims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func NewService(ctx context.Context, issuer string, jwksURL string) (*Service, error) {
	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("create jwks client: %w", err)
	}

	issuer = normalizeIssuer(issuer)

	return &Service{
		keyfunc: jwks.Keyfunc,
		issuer:  issuer,
	}, nil
}

func (s *Service) ValidateAccessToken(token string) (*UserClaims, error) {
	claims := &UserClaims{}

	parsed, err := jwt.ParseWithClaims(
		token,
		claims,
		s.keyfunc,
		jwt.WithIssuer(s.issuer),
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"}),
	)
	if err != nil {
		return nil, fmt.Errorf("parse supabase jwt: %w", err)
	}

	if !parsed.Valid {
		return nil, fmt.Errorf("supabase jwt is invalid")
	}

	if claims.Subject == "" {
		return nil, fmt.Errorf("supabase jwt subject missing")
	}

	return claims, nil
}

func normalizeIssuer(issuer string) string {
	issuer = strings.TrimSpace(issuer)
	issuer = strings.TrimRight(issuer, "/")

	if strings.HasSuffix(issuer, "/auth/v1") {
		return issuer
	}

	return issuer + "/auth/v1"
}
