package config

import (
	"reflect"
	"testing"
)

func TestAllowedCORSOriginsIncludesPublicAppURLAndExplicitOrigins(t *testing.T) {
	cfg := Config{
		PublicAppURL:       "https://www.getdecree.com/",
		CORSAllowedOrigins: []string{"https://getdecree.com", "https://www.getdecree.com"},
	}

	got := cfg.AllowedCORSOrigins()
	want := []string{
		"http://localhost:3000",
		"https://www.getdecree.com",
		"https://getdecree.com",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("AllowedCORSOrigins() = %#v, want %#v", got, want)
	}
}

func TestSplitCSVTrimsEmptyValues(t *testing.T) {
	got := splitCSV(" https://www.getdecree.com, ,https://getdecree.com ")
	want := []string{"https://www.getdecree.com", "https://getdecree.com"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitCSV() = %#v, want %#v", got, want)
	}
}
