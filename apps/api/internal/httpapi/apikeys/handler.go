package apikeys

import (
	"encoding/json"
	"errors"
	"net/http"

	identitypkg "humangate/internal/identity"
	"humangate/internal/identity/apikeys"
	"humangate/internal/identity/userctx"
	"humangate/internal/platform/pgxutil"
)

type Handler struct {
	service *apikeys.Service
}

func NewHandler(service *apikeys.Service) *Handler {
	return &Handler{service: service}
}

type createAPIKeyRequest struct {
	WorkspaceID string   `json:"workspace_id"`
	Name        string   `json:"name"`
	Scopes      []string `json:"scopes"`
}

type createAPIKeyResponse struct {
	APIKey struct {
		ID          string   `json:"id"`
		WorkspaceID string   `json:"workspace_id"`
		Name        string   `json:"name"`
		Prefix      string   `json:"prefix"`
		Scopes      []string `json:"scopes"`
		CreatedAt   string   `json:"created_at"`
		Secret      string   `json:"secret"`
	} `json:"api_key"`
}

func (h *Handler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	user, ok := userctx.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	var req createAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "The request body must be valid JSON.")
		return
	}

	result, err := h.service.CreateAPIKey(r.Context(), apikeys.CreateAPIKeyCommand{
		WorkspaceID: req.WorkspaceID,
		UserID:      user.UserID,
		Name:        req.Name,
		Scopes:      req.Scopes,
	})
	if err != nil {
		switch {
		case errors.Is(err, apikeys.ErrInvalidAPIKeyRequest):
			writeError(w, http.StatusBadRequest, "invalid_request", "The API key request is missing required fields.")
		case errors.Is(err, identitypkg.ErrInvalidUserID), errors.Is(err, identitypkg.ErrInvalidWorkspaceID):
			writeError(w, http.StatusBadRequest, "invalid_identifier", "The provided identifier is invalid.")
		case errors.Is(err, identitypkg.ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden", "You do not have permission to create API keys for this workspace.")
		default:
			writeError(w, http.StatusInternalServerError, "internal_error", "The API key could not be created.")
		}
		return
	}

	var resp createAPIKeyResponse
	resp.APIKey.ID = pgxutil.UUIDString(result.APIKey.ID)
	resp.APIKey.WorkspaceID = pgxutil.UUIDString(result.APIKey.WorkspaceID)
	resp.APIKey.Name = result.APIKey.Name
	resp.APIKey.Prefix = result.APIKey.KeyPrefix
	resp.APIKey.Scopes = result.APIKey.Scopes
	resp.APIKey.CreatedAt = result.APIKey.CreatedAt.Time.UTC().Format(http.TimeFormat)
	resp.APIKey.Secret = result.RawKey

	writeJSON(w, http.StatusCreated, resp)
}

func writeError(w http.ResponseWriter, statusCode int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
