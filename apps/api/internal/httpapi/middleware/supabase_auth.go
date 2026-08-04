package middleware

import (
	"net/http"
	"strings"

	"greenpost/internal/identity"
	"greenpost/internal/identity/supabaseauth"
	"greenpost/internal/identity/userctx"
)

func SupabaseAuth(auth *supabaseauth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				http.Error(w, identity.ErrUnauthorized.Error(), http.StatusUnauthorized)
				return
			}

			claims, err := auth.ValidateAccessToken(token)
			if err != nil {
				http.Error(w, identity.ErrUnauthorized.Error(), http.StatusUnauthorized)
				return
			}

			ctx := userctx.WithUser(r.Context(), userctx.User{
				UserID: claims.Subject,
				Email:  claims.Email,
				Role:   claims.Role,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) string {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" {
		return ""
	}

	const prefix = "Bearer "
	if strings.HasPrefix(authHeader, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
	}

	return ""
}
