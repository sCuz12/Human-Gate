package securitytest

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"greenpost/internal/approval"
)

func TestTenantQueriesScopeWorkspaceOwnedResources(t *testing.T) {
	files := []string{
		"approval_requests.sql",
		"approval_decisions.sql",
		"api_keys.sql",
		"audit_events.sql",
		"continuation_targets.sql",
		"decision_deliveries.sql",
		"policies.sql",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			queries := namedQueries(t, filepath.Join("..", "..", "db", "queries", file))
			for name, sql := range queries {
				if isGlobalWorkerQuery(name) {
					continue
				}
				if !strings.Contains(normalizeSQL(sql), "workspace_id") {
					t.Fatalf("%s must scope tenant-owned access by workspace_id:\n%s", name, sql)
				}
			}
		})
	}
}

func TestApprovalResourceQueriesRequireWorkspaceAndResourceID(t *testing.T) {
	tests := []struct {
		file     string
		query    string
		required []string
	}{
		{
			file:  "approval_requests.sql",
			query: "GetApprovalRequestByID",
			required: []string{
				"where workspace_id = $1",
				"and id = $2",
			},
		},
		{
			file:  "approval_requests.sql",
			query: "LockApprovalRequestForDecision",
			required: []string{
				"where workspace_id = $1",
				"and id = $2",
				"for update",
			},
		},
		{
			file:  "approval_decisions.sql",
			query: "GetApprovalDecisionByRequestID",
			required: []string{
				"where workspace_id = $1",
				"and approval_request_id = $2",
			},
		},
		{
			file:  "continuation_targets.sql",
			query: "GetContinuationTargetByRequestID",
			required: []string{
				"where workspace_id = $1",
				"and approval_request_id = $2",
			},
		},
		{
			file:  "audit_events.sql",
			query: "ListAuditEventsByRequestID",
			required: []string{
				"where workspace_id = $1",
				"and approval_request_id = $2",
			},
		},
		{
			file:  "decision_deliveries.sql",
			query: "GetDecisionDeliveryByApprovalRequestID",
			required: []string{
				"where dd.workspace_id = $1",
				"and ad.approval_request_id = $2",
				"ad.workspace_id = dd.workspace_id",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			query := queryByName(t, test.file, test.query)
			normalized := normalizeSQL(query)
			for _, required := range test.required {
				if !strings.Contains(normalized, normalizeSQL(required)) {
					t.Fatalf("%s missing required tenant isolation fragment %q:\n%s", test.query, required, query)
				}
			}
		})
	}
}

func TestIdempotencyKeyIsScopedByWorkspace(t *testing.T) {
	query := queryByName(t, "approval_requests.sql", "GetApprovalRequestByIdempotencyKey")
	for _, required := range []string{
		"where workspace_id = $1",
		"and idempotency_key = $2",
	} {
		if !strings.Contains(normalizeSQL(query), normalizeSQL(required)) {
			t.Fatalf("idempotency lookup missing %q:\n%s", required, query)
		}
	}

	service := readFile(t, filepath.Join("..", "approval", "service.go"))
	if !strings.Contains(service, "GetApprovalRequestByIdempotencyKeyParams{\n\t\tWorkspaceID:") {
		t.Fatal("SubmitApprovalRequest must pass workspace_id into idempotency lookup")
	}
	if !strings.Contains(service, "CreateApprovalRequestParams{\n\t\tWorkspaceID:") {
		t.Fatal("SubmitApprovalRequest must persist approval requests with workspace_id")
	}
}

func TestApprovalServiceChecksMembershipBeforeWorkspaceScopedReads(t *testing.T) {
	source := readFile(t, filepath.Join("..", "approval", "queries.go"))
	tests := []struct {
		method       string
		resourceCall string
	}{
		{method: "GetApprovalRequest", resourceCall: "GetApprovalRequestByID"},
		{method: "GetApprovalRequestDelivery", resourceCall: "GetDecisionDeliveryByApprovalRequestID"},
		{method: "ListApprovalRequestAuditEvents", resourceCall: "ListAuditEventsByRequestID"},
		{method: "ListApprovalRequests", resourceCall: "ListApprovalRequestsByWorkspace"},
	}

	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			body := functionBody(t, source, "func (s *Service) "+test.method)
			memberIndex := strings.Index(body, "GetWorkspaceMember")
			resourceIndex := strings.Index(body, test.resourceCall)
			if memberIndex == -1 {
				t.Fatalf("%s must verify workspace membership", test.method)
			}
			if resourceIndex == -1 {
				t.Fatalf("%s must call %s", test.method, test.resourceCall)
			}
			if memberIndex > resourceIndex {
				t.Fatalf("%s must verify workspace membership before %s", test.method, test.resourceCall)
			}
		})
	}
}

func TestDecisionServiceChecksMembershipBeforeLockingRequest(t *testing.T) {
	source := readFile(t, filepath.Join("..", "approval", "decisions.go"))
	body := functionBody(t, source, "func (s *Service) DecideApprovalRequest")

	memberIndex := strings.Index(body, "GetWorkspaceMember")
	lockIndex := strings.Index(body, "LockApprovalRequestForDecision")
	createIndex := strings.Index(body, "CreateApprovalDecision")
	if memberIndex == -1 || lockIndex == -1 || createIndex == -1 {
		t.Fatal("decision service must verify membership, lock request, and create decision")
	}
	if !(memberIndex < lockIndex && lockIndex < createIndex) {
		t.Fatal("decision service must verify membership, then lock the workspace-scoped request, before creating a decision")
	}
}

func TestCreateApprovalRequestUsesAuthenticatedAPIKeyWorkspace(t *testing.T) {
	source := readFile(t, filepath.Join("..", "..", "apps", "api", "internal", "httpapi", "approvals", "handler.go"))
	body := functionBody(t, source, "func (h *Handler) CreateApprovalRequest")

	if strings.Contains(body, "WorkspaceID: req.") {
		t.Fatal("create approval request must not trust a client-supplied workspace_id")
	}
	if !strings.Contains(body, "WorkspaceID:    pgxutil.UUIDString(authKey.Workspace.ID)") {
		t.Fatal("create approval request must derive workspace_id from the authenticated API key")
	}
}

func TestSchemaKeepsSecurityBackstopConstraints(t *testing.T) {
	migration := readFile(t, filepath.Join("..", "..", "supabase", "migrations", "20260728160000_create_core_tables.sql"))
	normalized := normalizeSQL(migration)

	required := map[string]string{
		"approval idempotency is workspace scoped":     "unique (workspace_id, idempotency_key)",
		"one final decision per approval request":      "unique (approval_request_id)",
		"one continuation target per approval request": "create table public.continuation_targets",
		"one delivery per decision":                    "unique (decision_id)",
		"raw api keys are not modeled":                 "key_hash text not null",
	}

	for name, fragment := range required {
		t.Run(name, func(t *testing.T) {
			if !strings.Contains(normalized, normalizeSQL(fragment)) {
				t.Fatalf("schema missing required security fragment %q", fragment)
			}
		})
	}

	if strings.Contains(normalized, "raw_key") || strings.Contains(normalized, "api_key text") {
		t.Fatal("schema must not contain raw API key storage")
	}

	tableConstraints := map[string]string{
		"approval_decisions":   "unique (approval_request_id)",
		"continuation_targets": "unique (approval_request_id)",
		"decision_deliveries":  "unique (decision_id)",
	}
	for table, constraint := range tableConstraints {
		t.Run(table+" backstop", func(t *testing.T) {
			section := tableDefinition(t, migration, table)
			if !strings.Contains(normalizeSQL(section), normalizeSQL(constraint)) {
				t.Fatalf("%s missing required constraint %q", table, constraint)
			}
		})
	}
}

func TestPublicApprovalSummariesDoNotExposeContinuationURLs(t *testing.T) {
	assertNoSensitiveJSONField(t, reflect.TypeOf(approval.ApprovalRequestSummary{}))
	assertNoSensitiveJSONField(t, reflect.TypeOf(approval.DeliverySummary{}))
	assertNoSensitiveJSONField(t, reflect.TypeOf(approval.AuditEventSummary{}))
}

func TestDecisionServiceGuardsExpiryBeforeCreatingDecision(t *testing.T) {
	source := readFile(t, filepath.Join("..", "approval", "decisions.go"))

	lockIndex := strings.Index(source, "LockApprovalRequestForDecision")
	expiryIndex := strings.Index(source, "request.ExpiresAt.Valid")
	createIndex := strings.Index(source, "CreateApprovalDecision")
	if lockIndex == -1 || expiryIndex == -1 || createIndex == -1 {
		t.Fatal("decision service must lock the request, check expiry, and create a decision")
	}
	if !(lockIndex < expiryIndex && expiryIndex < createIndex) {
		t.Fatal("decision service must lock and check expiry before creating a final decision")
	}
}

func TestAuditMetadataConstructionAvoidsSensitiveKeys(t *testing.T) {
	for _, path := range []string{
		filepath.Join("..", "approval", "service.go"),
		filepath.Join("..", "approval", "decisions.go"),
		filepath.Join("..", "approval", "expiry.go"),
	} {
		t.Run(filepath.Base(path), func(t *testing.T) {
			source := readFile(t, path)
			for _, sensitive := range []string{`"destination"`, `"url"`, `"continuation_url"`, `"raw_key"`, `"key_hash"`, `"token"`, `"secret"`} {
				if strings.Contains(source, sensitive) {
					t.Fatalf("audit-producing source contains sensitive metadata key %s", sensitive)
				}
			}
		})
	}
}

func namedQueries(t *testing.T, path string) map[string]string {
	t.Helper()
	content := readFile(t, path)
	re := regexp.MustCompile(`(?m)^-- name: ([A-Za-z0-9_]+) :(?:one|many|exec)\n`)
	matches := re.FindAllStringSubmatchIndex(content, -1)
	if len(matches) == 0 {
		t.Fatalf("no sqlc named queries found in %s", path)
	}

	queries := make(map[string]string, len(matches))
	for i, match := range matches {
		name := content[match[2]:match[3]]
		start := match[1]
		end := len(content)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		queries[name] = strings.TrimSpace(content[start:end])
	}
	return queries
}

func queryByName(t *testing.T, file string, name string) string {
	t.Helper()
	queries := namedQueries(t, filepath.Join("..", "..", "db", "queries", file))
	query, ok := queries[name]
	if !ok {
		t.Fatalf("missing query %s in %s", name, file)
	}
	return query
}

func isGlobalWorkerQuery(name string) bool {
	switch name {
	case "GetActiveAPIKeyByPrefixGlobal", "ListDueDecisionDeliveries", "ListDueExpiredApprovalRequests":
		return true
	default:
		return false
	}
}

func functionBody(t *testing.T, source string, signaturePrefix string) string {
	t.Helper()
	start := strings.Index(source, signaturePrefix)
	if start == -1 {
		t.Fatalf("missing function with prefix %q", signaturePrefix)
	}
	open := strings.Index(source[start:], "{")
	if open == -1 {
		t.Fatalf("missing function body for %q", signaturePrefix)
	}
	bodyStart := start + open
	depth := 0
	for i := bodyStart; i < len(source); i++ {
		switch source[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return source[bodyStart : i+1]
			}
		}
	}
	t.Fatalf("unterminated function body for %q", signaturePrefix)
	return ""
}

func assertNoSensitiveJSONField(t *testing.T, typ reflect.Type) {
	t.Helper()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonName == "" || jsonName == "-" {
			continue
		}
		switch strings.ToLower(jsonName) {
		case "destination", "url", "continuation_url", "resume_url", "webhook_url", "secret", "token", "key_hash", "raw_key":
			t.Fatalf("%s exposes sensitive json field %q", typ.Name(), jsonName)
		}
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
}

func tableDefinition(t *testing.T, migration string, table string) string {
	t.Helper()
	startMarker := "create table public." + table + " ("
	start := strings.Index(migration, startMarker)
	if start == -1 {
		t.Fatalf("missing table definition for %s", table)
	}
	remaining := migration[start:]
	end := strings.Index(remaining, "\n);")
	if end == -1 {
		t.Fatalf("unterminated table definition for %s", table)
	}
	return remaining[:end]
}

func normalizeSQL(sql string) string {
	return strings.Join(strings.Fields(strings.ToLower(sql)), " ")
}
