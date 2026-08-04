package delivery

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"humangate/db/generated"
	"humangate/internal/platform/pgxutil"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestSignedPayloadContainsDecisionIntegrityFields(t *testing.T) {
	service := NewService(nil, nil, nil, Config{SigningKey: "test-signing-key"})
	delivery := testDeliveryRow()

	body, err := service.signedPayload(delivery)
	if err != nil {
		t.Fatalf("signed payload: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	required := []string{
		"delivery_id",
		"workspace_id",
		"approval_request_id",
		"decision_id",
		"decision",
		"original_action_hash",
		"approved_action_hash",
		"issued_at",
		"signature_algorithm",
		"signature_header",
	}
	for _, key := range required {
		if _, ok := payload[key]; !ok {
			t.Fatalf("signed payload missing %q: %s", key, string(body))
		}
	}

	for _, forbidden := range []string{"destination", "url", "continuation_url", "webhook_url", "secret", "token"} {
		if _, ok := payload[forbidden]; ok {
			t.Fatalf("signed callback payload exposes %q: %s", forbidden, string(body))
		}
	}

	if payload["signature_algorithm"] != "HMAC-SHA256" {
		t.Fatalf("signature algorithm mismatch: got %v", payload["signature_algorithm"])
	}
	if payload["signature_header"] != "X-HumanGate-Signature" {
		t.Fatalf("signature header mismatch: got %v", payload["signature_header"])
	}
}

func TestSignUsesHMACSHA256(t *testing.T) {
	body := []byte(`{"decision_id":"decision-1","decision":"approved"}`)
	key := []byte("test-signing-key")

	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(body)
	want := hex.EncodeToString(mac.Sum(nil))

	if got := sign(body, key); got != want {
		t.Fatalf("signature mismatch\n got: %s\nwant: %s", got, want)
	}
}

func TestDeliverPostsSignedDecisionPayload(t *testing.T) {
	delivery := testDeliveryRow()
	service := NewService(nil, nil, nil, Config{SigningKey: "test-signing-key"})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method mismatch: got %s, want %s", r.Method, http.MethodPost)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("content type mismatch: got %q", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("X-HumanGate-Signature") == "" {
			t.Fatal("missing signature header")
		}
		if r.Header.Get("X-HumanGate-Decision-ID") == "" {
			t.Fatal("missing decision id header")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	delivery.Destination = pgtype.Text{String: server.URL, Valid: true}

	statusCode, err := service.deliver(t.Context(), delivery)
	if err != nil {
		t.Fatalf("deliver: %v", err)
	}
	if statusCode != http.StatusNoContent {
		t.Fatalf("status code mismatch: got %d, want %d", statusCode, http.StatusNoContent)
	}
}

func testDeliveryRow() generated.ListDueDecisionDeliveriesRow {
	issuedAt := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	return generated.ListDueDecisionDeliveriesRow{
		DeliveryID:           mustUUID("c655cc36-9bb6-4f6a-b509-12554c76fda1"),
		WorkspaceID:          mustUUID("0f480062-c549-4277-8f60-b42f38a263d9"),
		DecisionID:           mustUUID("3e94daca-1818-4c2a-8556-94992643e921"),
		ContinuationTargetID: mustUUID("a992b438-b0c4-4f93-b968-a419fbceff6f"),
		Status:               generated.DeliveryStatusPending,
		AttemptCount:         0,
		ApprovalRequestID:    mustUUID("09e85fc9-51c8-49f6-a232-3ae78a1427b0"),
		Decision:             generated.DecisionTypeApproved,
		OriginalActionHash:   "original-hash",
		ApprovedAction:       []byte(`{"type":"customer.refund","parameters":{"amount":82}}`),
		ApprovedActionHash:   pgtype.Text{String: "approved-hash", Valid: true},
		IssuedAt:             pgxutil.Timestamptz(issuedAt),
		OriginalAction:       []byte(`{"type":"customer.refund","parameters":{"amount":82}}`),
		ActionType:           "customer.refund",
		Strategy:             generated.ContinuationStrategyWebhook,
		Platform:             "n8n",
		Destination:          pgtype.Text{String: "https://workflow.example/resume/secret", Valid: true},
	}
}

func mustUUID(value string) pgtype.UUID {
	parsed, err := uuid.Parse(value)
	if err != nil {
		panic(err)
	}
	return pgxutil.UUID(parsed)
}
