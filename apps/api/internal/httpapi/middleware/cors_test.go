package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSAllowsConfiguredPreflightOrigin(t *testing.T) {
	handler := CORS([]string{"https://www.getdecree.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("preflight should not call next handler")
	}))

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/workspaces", nil)
	request.Header.Set("Origin", "https://www.getdecree.com")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	request.Header.Set("Access-Control-Request-Headers", "authorization")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "https://www.getdecree.com" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, "https://www.getdecree.com")
	}
}

func TestCORSDoesNotAllowUnconfiguredOrigin(t *testing.T) {
	handler := CORS([]string{"https://www.getdecree.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("preflight should not call next handler")
	}))

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/workspaces", nil)
	request.Header.Set("Origin", "https://attacker.example")

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}
