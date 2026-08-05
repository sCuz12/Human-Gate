package policies

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"decree/internal/identity"
	"decree/internal/identity/userctx"
	"decree/internal/policy"
)

type Handler struct {
	service *policy.Service
	logger  *slog.Logger
}

type createPolicyRequest struct {
	WorkspaceID     string             `json:"workspace_id"`
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	Priority        int32              `json:"priority"`
	IsActive        bool               `json:"is_active"`
	Conditions      []policy.Condition `json:"conditions"`
	Effect          string             `json:"effect"`
	DeadlineSeconds int64              `json:"deadline_seconds"`
}

type updatePolicyRequest = createPolicyRequest

func NewHandler(service *policy.Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	user, ok := userctx.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	policies, err := h.service.ListPolicies(
		r.Context(),
		strings.TrimSpace(r.URL.Query().Get("workspace_id")),
		user.UserID,
	)
	if err != nil {
		h.writePolicyError(r, w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"policies": policies})
}

func (h *Handler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	user, ok := userctx.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	var req createPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "The request body must be valid JSON.")
		return
	}

	createdPolicy, err := h.service.CreatePolicy(r.Context(), policy.CreatePolicyCommand{
		WorkspaceID:     req.WorkspaceID,
		UserID:          user.UserID,
		Name:            req.Name,
		Description:     req.Description,
		Priority:        req.Priority,
		IsActive:        req.IsActive,
		Conditions:      req.Conditions,
		Effect:          req.Effect,
		DeadlineSeconds: req.DeadlineSeconds,
	})
	if err != nil {
		h.writePolicyError(r, w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"policy": createdPolicy})
}

func (h *Handler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	user, ok := userctx.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	var req updatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "The request body must be valid JSON.")
		return
	}

	updatedPolicy, err := h.service.UpdatePolicy(r.Context(), policy.UpdatePolicyCommand{
		WorkspaceID:     req.WorkspaceID,
		PolicyID:        strings.TrimSpace(chi.URLParam(r, "id")),
		UserID:          user.UserID,
		Name:            req.Name,
		Description:     req.Description,
		Priority:        req.Priority,
		IsActive:        req.IsActive,
		Conditions:      req.Conditions,
		Effect:          req.Effect,
		DeadlineSeconds: req.DeadlineSeconds,
	})
	if err != nil {
		h.writePolicyError(r, w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"policy": updatedPolicy})
}

func (h *Handler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	user, ok := userctx.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.")
		return
	}

	err := h.service.DeletePolicy(r.Context(), policy.DeletePolicyCommand{
		WorkspaceID: strings.TrimSpace(r.URL.Query().Get("workspace_id")),
		PolicyID:    strings.TrimSpace(chi.URLParam(r, "id")),
		UserID:      user.UserID,
	})
	if err != nil {
		h.writePolicyError(r, w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) writePolicyError(r *http.Request, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, policy.ErrInvalidRequest), errors.Is(err, identity.ErrInvalidWorkspaceID), errors.Is(err, identity.ErrInvalidUserID):
		writeError(w, http.StatusBadRequest, "invalid_request", "The policy request is invalid.")
	case errors.Is(err, policy.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden", "You cannot manage policies in this workspace.")
	case errors.Is(err, policy.ErrNotFound):
		writeError(w, http.StatusNotFound, "policy_not_found", "The policy could not be found.")
	default:
		h.logger.ErrorContext(r.Context(), "policy request failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Policies could not be processed.")
	}
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
