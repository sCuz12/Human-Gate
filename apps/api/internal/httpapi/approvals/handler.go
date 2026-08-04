package approvals

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"greenpost/internal/approval"
	"greenpost/internal/identity/apikeys"
	"greenpost/internal/identity/userctx"
	"greenpost/internal/platform/pgxutil"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	apiKeys   *apikeys.Service
	approvals *approval.Service
	logger    *slog.Logger
}

func NewHandler(apiKeys *apikeys.Service, approvals *approval.Service, logger *slog.Logger) *Handler {
	return &Handler{
		apiKeys:   apiKeys,
		approvals: approvals,
		logger:    logger,
	}
}

type createApprovalRequestRequest struct {
	Action struct {
		Type        string          `json:"type"`
		Title       string          `json:"title"`
		Description string          `json:"description"`
		Parameters  json.RawMessage `json:"parameters"`
	} `json:"action"`
	Context struct {
		Reason          string            `json:"reason"`
		Evidence        json.RawMessage   `json:"evidence"`
		Confidence      *float64          `json:"confidence"`
		Reversible      *bool             `json:"reversible"`
		AffectedSystems []string          `json:"affected_systems"`
		Deadline        string            `json:"deadline"`
		Metadata        map[string]string `json:"metadata"`
	} `json:"context"`
	Source struct {
		Platform    string `json:"platform"`
		WorkflowID  string `json:"workflow_id"`
		ExecutionID string `json:"execution_id"`
	} `json:"source"`
	Continuation struct {
		Strategy string `json:"strategy"`
		URL      string `json:"url"`
	} `json:"continuation"`
}

type errorResponse struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id,omitempty"`
	} `json:"error"`
}

type createApprovalRequestResponse struct {
	RequestID        string `json:"request_id"`
	Status           string `json:"status"`
	DecisionRequired bool   `json:"decision_required"`
}

type listApprovalRequestsResponse struct {
	Requests []approval.ApprovalRequestSummary `json:"requests"`
}

type getApprovalRequestResponse struct {
	Request approval.ApprovalRequestSummary `json:"request"`
}

type getApprovalRequestDeliveryResponse struct {
	Delivery approval.DeliverySummary `json:"delivery"`
}

type listApprovalRequestAuditEventsResponse struct {
	AuditEvents []approval.AuditEventSummary `json:"audit_events"`
}

type decideApprovalRequestRequest struct {
	WorkspaceID string `json:"workspace_id"`
	Comment     string `json:"comment"`
}

func (h *Handler) CreateApprovalRequest(w http.ResponseWriter, r *http.Request) {
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		writeError(w, http.StatusBadRequest, "missing_idempotency_key", "Idempotency-Key header is required.", "")
		return
	}

	authKey, err := h.apiKeys.Authenticate(r.Context(), bearerToken(r), "approval_requests:create")
	if err != nil {
		switch {
		case errors.Is(err, apikeys.ErrMissingAPIKey), errors.Is(err, apikeys.ErrInvalidAPIKey):
			writeError(w, http.StatusUnauthorized, "invalid_api_key", "A valid API key is required.", "")
		case errors.Is(err, apikeys.ErrAPIKeyUnauthorized):
			writeError(w, http.StatusForbidden, "insufficient_scope", "The API key does not have the required scope.", "")
		default:
			h.logger.ErrorContext(r.Context(), "api key authentication failed", "error", err, "path", r.URL.Path)
			writeError(w, http.StatusInternalServerError, "internal_error", "The request could not be authenticated.", "")
		}
		return
	}

	var req createApprovalRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "The request body must be valid JSON.", "")
		return
	}

	result, err := h.approvals.SubmitApprovalRequest(r.Context(), approval.SubmitApprovalRequestCommand{
		WorkspaceID:    pgxutil.UUIDString(authKey.Workspace.ID),
		IdempotencyKey: idempotencyKey,
		Action: approval.ProposedAction{
			Type:        req.Action.Type,
			Title:       req.Action.Title,
			Description: req.Action.Description,
			Parameters:  req.Action.Parameters,
		},
		Context: approval.RequestContext{
			Reason:          req.Context.Reason,
			Evidence:        req.Context.Evidence,
			Confidence:      req.Context.Confidence,
			Reversible:      req.Context.Reversible,
			AffectedSystems: req.Context.AffectedSystems,
			Deadline:        req.Context.Deadline,
			Metadata:        req.Context.Metadata,
		},
		Source: approval.RequestSource{
			Platform:    req.Source.Platform,
			WorkflowID:  req.Source.WorkflowID,
			ExecutionID: req.Source.ExecutionID,
		},
		Continuation: approval.Continuation{
			Strategy: req.Continuation.Strategy,
			URL:      req.Continuation.URL,
		},
		AuthenticatedBy: approval.AuthenticatedActor{
			ActorType: "api_key",
			ActorID:   pgxutil.UUIDString(authKey.APIKey.ID),
		},
	})
	if err != nil {
		switch {
		case errors.Is(err, approval.ErrInvalidRequest):
			writeError(w, http.StatusBadRequest, "invalid_request", "The approval request is missing required fields.", "")
		case errors.Is(err, pgx.ErrNoRows):
			writeError(w, http.StatusNotFound, "workspace_not_found", "The workspace could not be found.", "")
		default:
			h.logger.ErrorContext(
				r.Context(),
				"create approval request failed",
				"error", err,
				"workspace_id", pgxutil.UUIDString(authKey.Workspace.ID),
				"action_type", req.Action.Type,
				"source_platform", req.Source.Platform,
			)
			writeError(w, http.StatusInternalServerError, "internal_error", "The approval request could not be created.", "")
		}
		return
	}

	statusCode := http.StatusCreated
	if result.Existing {
		statusCode = http.StatusOK
	}

	writeJSON(w, statusCode, createApprovalRequestResponse{
		RequestID:        pgxutil.UUIDString(result.Request.ID),
		Status:           string(result.Request.Status),
		DecisionRequired: result.Request.DecisionRequired,
	})
}

func (h *Handler) ListApprovalRequests(w http.ResponseWriter, r *http.Request) {
	user, ok := userctx.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.", "")
		return
	}

	workspaceID := strings.TrimSpace(r.URL.Query().Get("workspace_id"))
	limit := int32(queryInt(r, "limit", 50))
	offset := int32(queryInt(r, "offset", 0))

	requests, err := h.approvals.ListApprovalRequests(r.Context(), approval.ListApprovalRequestsCommand{
		WorkspaceID: workspaceID,
		UserID:      user.UserID,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		switch {
		case errors.Is(err, approval.ErrInvalidRequest):
			writeError(w, http.StatusBadRequest, "invalid_request", "workspace_id is required.", "")
		case errors.Is(err, approval.ErrForbidden):
			writeError(w, http.StatusForbidden, "forbidden", "You do not have access to this workspace.", "")
		default:
			h.logger.ErrorContext(
				r.Context(),
				"list approval requests failed",
				"error", err,
				"workspace_id", workspaceID,
				"user_id", user.UserID,
			)
			writeError(w, http.StatusInternalServerError, "internal_error", "Approval requests could not be loaded.", "")
		}
		return
	}

	writeJSON(w, http.StatusOK, listApprovalRequestsResponse{Requests: requests})
}

func (h *Handler) GetApprovalRequest(w http.ResponseWriter, r *http.Request) {
	user, ok := userctx.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.", "")
		return
	}

	request, err := h.approvals.GetApprovalRequest(r.Context(), approval.GetApprovalRequestCommand{
		WorkspaceID: strings.TrimSpace(r.URL.Query().Get("workspace_id")),
		RequestID:   strings.TrimSpace(chi.URLParam(r, "id")),
		UserID:      user.UserID,
	})
	if err != nil {
		h.writeApprovalError(r, w, err)
		return
	}

	writeJSON(w, http.StatusOK, getApprovalRequestResponse{Request: request})
}

func (h *Handler) GetApprovalRequestDelivery(w http.ResponseWriter, r *http.Request) {
	user, ok := userctx.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.", "")
		return
	}

	delivery, err := h.approvals.GetApprovalRequestDelivery(r.Context(), approval.GetApprovalRequestDeliveryCommand{
		WorkspaceID: strings.TrimSpace(r.URL.Query().Get("workspace_id")),
		RequestID:   strings.TrimSpace(chi.URLParam(r, "id")),
		UserID:      user.UserID,
	})
	if err != nil {
		h.writeApprovalError(r, w, err)
		return
	}

	writeJSON(w, http.StatusOK, getApprovalRequestDeliveryResponse{Delivery: delivery})
}

func (h *Handler) ListApprovalRequestAuditEvents(w http.ResponseWriter, r *http.Request) {
	user, ok := userctx.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.", "")
		return
	}

	events, err := h.approvals.ListApprovalRequestAuditEvents(r.Context(), approval.ListApprovalRequestAuditEventsCommand{
		WorkspaceID: strings.TrimSpace(r.URL.Query().Get("workspace_id")),
		RequestID:   strings.TrimSpace(chi.URLParam(r, "id")),
		UserID:      user.UserID,
	})
	if err != nil {
		h.writeApprovalError(r, w, err)
		return
	}

	writeJSON(w, http.StatusOK, listApprovalRequestAuditEventsResponse{AuditEvents: events})
}

func (h *Handler) ApproveApprovalRequest(w http.ResponseWriter, r *http.Request) {
	h.decideApprovalRequest(w, r, "approve")
}

func (h *Handler) RejectApprovalRequest(w http.ResponseWriter, r *http.Request) {
	h.decideApprovalRequest(w, r, "reject")
}

func (h *Handler) decideApprovalRequest(w http.ResponseWriter, r *http.Request, decision string) {
	user, ok := userctx.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication is required.", "")
		return
	}

	var req decideApprovalRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "The request body must be valid JSON.", "")
		return
	}

	request, err := h.approvals.DecideApprovalRequest(r.Context(), approval.DecisionCommand{
		WorkspaceID: req.WorkspaceID,
		RequestID:   strings.TrimSpace(chi.URLParam(r, "id")),
		UserID:      user.UserID,
		Decision:    decision,
		Comment:     req.Comment,
	})
	if err != nil {
		h.logger.ErrorContext(
			r.Context(),
			"approval decision failed",
			"error", err,
			"workspace_id", req.WorkspaceID,
			"approval_request_id", strings.TrimSpace(chi.URLParam(r, "id")),
			"user_id", user.UserID,
			"decision", decision,
		)
		h.writeApprovalError(r, w, err)
		return
	}

	writeJSON(w, http.StatusOK, getApprovalRequestResponse{Request: request})
}

func (h *Handler) writeApprovalError(r *http.Request, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, approval.ErrInvalidRequest):
		writeError(w, http.StatusBadRequest, "invalid_request", "The approval request input is invalid.", "")
	case errors.Is(err, approval.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden", "You do not have access to this request.", "")
	case errors.Is(err, approval.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "The approval request was not found.", "")
	case errors.Is(err, approval.ErrResolved):
		writeError(w, http.StatusConflict, "request_resolved", "This approval request has already been resolved.", "")
	case errors.Is(err, approval.ErrExpired):
		writeError(w, http.StatusConflict, "request_expired", "This approval request has expired.", "")
	default:
		h.logger.ErrorContext(
			r.Context(),
			"approval request operation failed",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path,
		)
		writeError(w, http.StatusInternalServerError, "internal_error", "The approval request could not be processed.", "")
	}
}

func bearerToken(r *http.Request) string {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" {
		return strings.TrimSpace(r.Header.Get("X-API-Key"))
	}

	const prefix = "Bearer "
	if strings.HasPrefix(authHeader, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
	}

	return authHeader
}

func queryInt(r *http.Request, key string, fallback int) int {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func writeError(w http.ResponseWriter, statusCode int, code string, message string, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	payload := errorResponse{}
	payload.Error.Code = code
	payload.Error.Message = message
	payload.Error.RequestID = requestID

	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
