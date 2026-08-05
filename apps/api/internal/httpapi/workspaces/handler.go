package workspaces

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	identitypkg "decree/internal/identity"
	"decree/internal/identity/userctx"
	identityworkspaces "decree/internal/identity/workspaces"
)

type Handler struct {
	service *identityworkspaces.Service
	logger  *slog.Logger
}

type createWorkspaceRequest struct {
	Name string `json:"name"`
}

func NewHandler(service *identityworkspaces.Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) CreateWorkspace(w http.ResponseWriter, r *http.Request) {
	user, ok := userctx.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	var req createWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "The request body must be valid JSON.")
		return
	}

	workspace, err := h.service.CreateWorkspace(r.Context(), identityworkspaces.CreateWorkspaceCommand{
		UserID: user.UserID,
		Name:   req.Name,
	})
	if err != nil {
		switch {
		case errors.Is(err, identityworkspaces.ErrInvalidWorkspaceRequest):
			writeError(w, http.StatusBadRequest, "invalid_request", "A workspace name is required.")
		case errors.Is(err, identitypkg.ErrInvalidUserID):
			writeError(w, http.StatusBadRequest, "invalid_user", "The authenticated user is invalid.")
		default:
			h.logger.ErrorContext(r.Context(), "create workspace failed", "error", err, "user_id", user.UserID)
			writeError(w, http.StatusInternalServerError, "internal_error", "The workspace could not be created.")
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"workspace": workspace})
}

func (h *Handler) ListWorkspaces(w http.ResponseWriter, r *http.Request) {
	user, ok := userctx.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	workspaces, err := h.service.ListWorkspaces(r.Context(), user.UserID)
	if err != nil {
		switch {
		case errors.Is(err, identityworkspaces.ErrInvalidWorkspaceRequest), errors.Is(err, identitypkg.ErrInvalidUserID):
			writeError(w, http.StatusBadRequest, "invalid_user", "The authenticated user is invalid.")
		default:
			h.logger.ErrorContext(r.Context(), "list workspaces failed", "error", err, "user_id", user.UserID)
			writeError(w, http.StatusInternalServerError, "internal_error", "Workspaces could not be loaded.")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"workspaces": workspaces})
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
