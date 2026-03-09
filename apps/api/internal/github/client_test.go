package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRepositoriesReconnectRequiredOnUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Bad credentials"}`))
	}))
	defer server.Close()

	client := NewHTTPGitHubClient("cid", "secret", "", server.URL)
	_, err := client.ListRepositories(context.Background(), "expired-token")
	if err == nil {
		t.Fatal("expected error")
	}
	if err != ErrGitHubReconnectRequired {
		t.Fatalf("expected ErrGitHubReconnectRequired, got %v", err)
	}
}

func TestListRepositoriesRateLimited(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"API rate limit exceeded"}`))
	}))
	defer server.Close()

	client := NewHTTPGitHubClient("cid", "secret", "", server.URL)
	_, err := client.ListRepositories(context.Background(), "token")
	if err == nil {
		t.Fatal("expected error")
	}
	if err != ErrGitHubRateLimited {
		t.Fatalf("expected ErrGitHubRateLimited, got %v", err)
	}
}
