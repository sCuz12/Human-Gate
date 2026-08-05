package middleware

import (
	"net/http"
	"strings"

	"decree/internal/identity"
	"decree/internal/identity/userctx"

	"github.com/google/uuid"
)

// DevUserAuth is a temporary bridge until real Supabase JWT validation is added.
func DevUserAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawUserID := strings.TrimSpace(r.Header.Get("X-User-ID"))
		if rawUserID == "" {
			http.Error(w, identity.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		if _, err := uuid.Parse(rawUserID); err != nil {
			http.Error(w, identity.ErrInvalidUserID.Error(), http.StatusUnauthorized)
			return
		}

		ctx := userctx.WithUser(r.Context(), userctx.User{UserID: rawUserID})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
