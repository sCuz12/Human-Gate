package apikeys

import (
	"strings"
	"testing"
)

func TestGenerateRawAPIKeyOnlyReturnsPrefixForStorage(t *testing.T) {
	rawKey, prefix, err := generateRawAPIKey()
	if err != nil {
		t.Fatalf("generate raw api key: %v", err)
	}

	if !strings.HasPrefix(rawKey, prefix+".") {
		t.Fatalf("raw key must include the returned display prefix; raw=%q prefix=%q", rawKey, prefix)
	}
	if strings.Contains(prefix, ".") {
		t.Fatalf("stored prefix must not contain secret material: %q", prefix)
	}
	if prefix == rawKey {
		t.Fatal("stored prefix must not equal the raw API key")
	}
}

func TestHashAPIKeyDoesNotStoreRawSecret(t *testing.T) {
	rawKey := "hgk_123456789abc.super-secret-test-key"
	hash := HashAPIKey(rawKey)

	if hash == rawKey {
		t.Fatal("hash must not equal the raw API key")
	}
	if strings.Contains(hash, "super-secret-test-key") || strings.Contains(hash, "hgk_123456789abc") {
		t.Fatalf("hash must not contain raw API key material: %q", hash)
	}
	if len(hash) != 64 {
		t.Fatalf("sha256 hex hash length mismatch: got %d, want 64", len(hash))
	}
}

func TestExtractPrefixRejectsKeysWithoutSecretSeparator(t *testing.T) {
	tests := []string{
		"",
		"hgk_123456789abc",
		".secret",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			if prefix, ok := extractPrefix(test); ok {
				t.Fatalf("expected invalid key, got prefix %q", prefix)
			}
		})
	}
}
