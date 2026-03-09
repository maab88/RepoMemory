package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteErrorIncludesRequestIDFromHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	rr.Header().Set("X-Request-Id", "req-abc")

	WriteError(rr, http.StatusBadRequest, "bad_request", "invalid payload")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}

	var body Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Error == nil || body.Error.RequestID != "req-abc" {
		t.Fatalf("expected requestId in error envelope, got %+v", body.Error)
	}
}

func TestWriteErrorGeneratesRequestIDWhenMissing(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteError(rr, http.StatusUnauthorized, "unauthorized", "missing current user")

	var body Envelope
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body.Error == nil || body.Error.RequestID == "" {
		t.Fatalf("expected generated requestId, got %+v", body.Error)
	}
}
