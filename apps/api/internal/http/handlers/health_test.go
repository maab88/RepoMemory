package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/maab88/repomemory/apps/api/internal/testutil"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	Health(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}

	got := testutil.DecodeJSON[HealthResponse](t, rr)

	if got.Status != "ok" {
		t.Fatalf("expected status ok, got %s", got.Status)
	}

	if got.Service != "api" {
		t.Fatalf("expected service api, got %s", got.Service)
	}

	if got.Timestamp == "" {
		t.Fatal("expected timestamp to be set")
	}
}
