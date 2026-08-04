package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/jackc/pgx/v5/pgxpool"
	"greenpost/apps/api/internal/httpapi/apikeys"
	"greenpost/apps/api/internal/httpapi/approvals"
	"greenpost/apps/api/internal/httpapi/health"
	"greenpost/apps/api/internal/httpapi/middleware"
	"greenpost/apps/api/internal/httpapi/policies"
	"greenpost/apps/api/internal/httpapi/workspaces"
	"greenpost/internal/approval"
	identityapikeys "greenpost/internal/identity/apikeys"
	"greenpost/internal/identity/supabaseauth"
	identityworkspaces "greenpost/internal/identity/workspaces"
	"greenpost/internal/policy"
)

func NewRouter(logger *slog.Logger, db *pgxpool.Pool, supabaseAuth *supabaseauth.Service, allowedOrigins []string) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer(logger))
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.CORS(allowedOrigins))
	approvalHandler := approvals.NewHandler(
		identityapikeys.NewService(db, nil),
		approval.NewService(db, nil, logger),
		logger,
	)
	apiKeyHandler := apikeys.NewHandler(identityapikeys.NewService(db, nil))
	workspaceHandler := workspaces.NewHandler(identityworkspaces.NewService(db), logger)
	policyHandler := policies.NewHandler(policy.NewService(db), logger)

	router.Get("/healthz", health.Handle())
	router.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", health.Handle())
		r.Post("/approval-requests", approvalHandler.CreateApprovalRequest)
		r.Group(func(r chi.Router) {
			r.Use(middleware.SupabaseAuth(supabaseAuth))
			r.Get("/approval-requests", approvalHandler.ListApprovalRequests)
			r.Get("/approval-requests/{id}", approvalHandler.GetApprovalRequest)
			r.Get("/approval-requests/{id}/audit-events", approvalHandler.ListApprovalRequestAuditEvents)
			r.Get("/approval-requests/{id}/delivery", approvalHandler.GetApprovalRequestDelivery)
			r.Post("/approval-requests/{id}/approve", approvalHandler.ApproveApprovalRequest)
			r.Post("/approval-requests/{id}/reject", approvalHandler.RejectApprovalRequest)
			r.Get("/workspaces", workspaceHandler.ListWorkspaces)
			r.Post("/workspaces", workspaceHandler.CreateWorkspace)
			r.Post("/api-keys", apiKeyHandler.CreateAPIKey)
			r.Get("/policies", policyHandler.ListPolicies)
			r.Post("/policies", policyHandler.CreatePolicy)
			r.Patch("/policies/{id}", policyHandler.UpdatePolicy)
			r.Delete("/policies/{id}", policyHandler.DeletePolicy)
		})
	})

	_ = logger

	return router
}
