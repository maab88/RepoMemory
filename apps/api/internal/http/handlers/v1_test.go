package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/auth"
	"github.com/maab88/repomemory/apps/api/internal/org"
)

type fakeResolver struct{}

func (f fakeResolver) Resolve(_ context.Context, input auth.MockUserInput) (auth.CurrentUser, error) {
	return auth.CurrentUser{ID: uuid.NewSHA1(uuid.NameSpaceOID, []byte(input.RawID)), DisplayName: "Dev Tester", Email: input.Email}, nil
}

type fakeOrgService struct {
	listResult   []org.OrganizationWithRole
	createResult org.OrganizationWithRole
	getResult    org.OrganizationWithRole
	createErr    error
	listErr      error
	getErr       error
}

func (f *fakeOrgService) CreateOrganization(_ context.Context, _ uuid.UUID, _ string) (org.OrganizationWithRole, error) {
	return f.createResult, f.createErr
}

func (f *fakeOrgService) ListOrganizations(_ context.Context, _ uuid.UUID) ([]org.OrganizationWithRole, error) {
	return f.listResult, f.listErr
}

func (f *fakeOrgService) GetOrganization(_ context.Context, _, _ uuid.UUID) (org.OrganizationWithRole, error) {
	return f.getResult, f.getErr
}

func newTestRouter(orgSvc OrganizationService) http.Handler {
	h := NewV1Handler(orgSvc)
	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/me", h.GetMe)
		r.Get("/organizations", h.ListOrganizations)
		r.Post("/organizations", h.CreateOrganization)
	})
	return r
}

func TestGetMeSuccess(t *testing.T) {
	router := newTestRouter(&fakeOrgService{})
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	req.Header.Set("x-user-id", "user-1")
	req.Header.Set("x-user-email", "user@example.com")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetOrganizationsSuccess(t *testing.T) {
	orgID := uuid.New()
	router := newTestRouter(&fakeOrgService{listResult: []org.OrganizationWithRole{{ID: orgID, Name: "Acme", Slug: "acme", Role: "owner"}}})
	req := httptest.NewRequest(http.MethodGet, "/v1/organizations", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var payload struct {
		Data []org.OrganizationWithRole `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if len(payload.Data) != 1 {
		t.Fatalf("expected 1 org, got %d", len(payload.Data))
	}
}

func TestPostOrganizationsSuccess(t *testing.T) {
	created := org.OrganizationWithRole{ID: uuid.New(), Name: "Acme", Slug: "acme", Role: "owner"}
	router := newTestRouter(&fakeOrgService{createResult: created})
	body := bytes.NewBufferString(`{"name":"Acme"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/organizations", body)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
}

func TestAuthMissing(t *testing.T) {
	router := newTestRouter(&fakeOrgService{})
	req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestPostOrganizationsBadPayload(t *testing.T) {
	router := newTestRouter(&fakeOrgService{})
	body := bytes.NewBufferString(`{`)
	req := httptest.NewRequest(http.MethodPost, "/v1/organizations", body)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
