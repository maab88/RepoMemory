package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

type stubIdentityValidator struct {
	identity ExternalIdentity
	err      error
}

func (s stubIdentityValidator) ValidateBearerToken(_ context.Context, _ string) (ExternalIdentity, error) {
	return s.identity, s.err
}

type stubIdentityMapper struct {
	user CurrentUser
	err  error
}

func (s stubIdentityMapper) MapToCurrentUser(_ context.Context, _ ExternalIdentity) (CurrentUser, error) {
	return s.user, s.err
}

func TestRequireBearerAuthRejectsMissingAuthorization(t *testing.T) {
	mw := RequireBearerAuth(stubIdentityValidator{}, stubIdentityMapper{})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestRequireBearerAuthRejectsInvalidToken(t *testing.T) {
	mw := RequireBearerAuth(stubIdentityValidator{err: ErrInvalidAuthToken}, stubIdentityMapper{})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestRequireBearerAuthInjectsCurrentUser(t *testing.T) {
	expected := CurrentUser{ID: uuid.New(), DisplayName: "Signed In User"}
	mw := RequireBearerAuth(
		stubIdentityValidator{identity: ExternalIdentity{Subject: "sub-1", Issuer: "issuer"}},
		stubIdentityMapper{user: expected},
	)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := CurrentUserFromContext(r.Context())
		if !ok {
			t.Fatal("expected current user in context")
		}
		if user.ID != expected.ID {
			t.Fatalf("expected %s, got %s", expected.ID, user.ID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
