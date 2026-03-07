package testutil

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func DecodeJSON[T any](t *testing.T, rr *httptest.ResponseRecorder) T {
	t.Helper()

	var out T
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return out
}
