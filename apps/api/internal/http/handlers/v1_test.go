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
	gh "github.com/maab88/repomemory/apps/api/internal/github"
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

type fakeGitHubOAuthService struct {
	redirectURL string
	account     gh.GitHubConnectionSummary
	startErr    error
	callbackErr error
	callbacks   int
}

func (f *fakeGitHubOAuthService) StartConnect(_ context.Context, _ gh.OAuthStartInput) (string, error) {
	return f.redirectURL, f.startErr
}

func (f *fakeGitHubOAuthService) HandleCallback(_ context.Context, _ gh.OAuthCallbackInput) (gh.GitHubConnectionSummary, error) {
	f.callbacks++
	return f.account, f.callbackErr
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
	return newTestRouterWithGitHub(orgSvc, &fakeGitHubOAuthService{})
}

func newTestRouterWithGitHub(orgSvc OrganizationService, githubSvc GitHubOAuthService) http.Handler {
	h := NewV1Handler(orgSvc, githubSvc)
	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/me", h.GetMe)
		r.Get("/organizations", h.ListOrganizations)
		r.Post("/organizations", h.CreateOrganization)
		r.Get("/organizations/{orgId}", h.GetOrganization)
		r.Post("/github/connect/start", h.StartGitHubConnect)
		r.Get("/github/callback", h.GitHubCallback)
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

func TestGetOrganizationSuccess(t *testing.T) {
	orgID := uuid.New()
	router := newTestRouter(&fakeOrgService{getResult: org.OrganizationWithRole{ID: orgID, Name: "Acme", Slug: "acme", Role: "owner"}})
	req := httptest.NewRequest(http.MethodGet, "/v1/organizations/"+orgID.String(), nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var payload struct {
		Data struct {
			ID   uuid.UUID `json:"id"`
			Name string    `json:"name"`
			Slug string    `json:"slug"`
			Role string    `json:"role"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if payload.Data.ID != orgID || payload.Data.Name != "Acme" || payload.Data.Slug != "acme" || payload.Data.Role != "owner" {
		t.Fatalf("unexpected organization payload: %+v", payload.Data)
	}
}

func TestGetOrganizationBadID(t *testing.T) {
	router := newTestRouter(&fakeOrgService{})
	req := httptest.NewRequest(http.MethodGet, "/v1/organizations/not-a-uuid", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestStartGitHubConnectReturnsRedirectURL(t *testing.T) {
	githubSvc := &fakeGitHubOAuthService{redirectURL: "https://github.com/login/oauth/authorize?state=abc"}
	router := newTestRouterWithGitHub(&fakeOrgService{}, githubSvc)

	req := httptest.NewRequest(http.MethodPost, "/v1/github/connect/start", bytes.NewBufferString(`{}`))
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var payload struct {
		Data struct {
			RedirectURL string `json:"redirectUrl"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Data.RedirectURL == "" {
		t.Fatal("expected redirectUrl to be present")
	}
}

func TestGitHubCallbackMissingCode(t *testing.T) {
	router := newTestRouterWithGitHub(&fakeOrgService{}, &fakeGitHubOAuthService{callbackErr: gh.ErrOAuthCodeMissing})
	req := httptest.NewRequest(http.MethodGet, "/v1/github/callback?state=abc", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGitHubCallbackInvalidState(t *testing.T) {
	router := newTestRouterWithGitHub(&fakeOrgService{}, &fakeGitHubOAuthService{callbackErr: gh.ErrInvalidState})
	req := httptest.NewRequest(http.MethodGet, "/v1/github/callback?state=bad&code=123", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGitHubCallbackTokenExchangeFailure(t *testing.T) {
	router := newTestRouterWithGitHub(&fakeOrgService{}, &fakeGitHubOAuthService{callbackErr: gh.ErrTokenExchangeFailed})
	req := httptest.NewRequest(http.MethodGet, "/v1/github/callback?state=ok&code=123", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", rr.Code)
	}
}

func TestGitHubCallbackSuccess(t *testing.T) {
	account := gh.GitHubConnectionSummary{
		ID:           uuid.New(),
		GitHubLogin:  "octocat",
		GitHubUserID: "12345",
	}
	githubSvc := &fakeGitHubOAuthService{account: account}
	router := newTestRouterWithGitHub(&fakeOrgService{}, githubSvc)
	req := httptest.NewRequest(http.MethodGet, "/v1/github/callback?state=ok&code=123", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var payload struct {
		Data struct {
			Connected bool `json:"connected"`
			Account   struct {
				GitHubLogin string `json:"githubLogin"`
			} `json:"account"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if !payload.Data.Connected || payload.Data.Account.GitHubLogin != "octocat" {
		t.Fatalf("unexpected callback payload: %+v", payload.Data)
	}
	if githubSvc.callbacks != 1 {
		t.Fatalf("expected callback service to be called once, got %d", githubSvc.callbacks)
	}
}

func TestStartGitHubConnectOrganizationForbidden(t *testing.T) {
	router := newTestRouterWithGitHub(&fakeOrgService{}, &fakeGitHubOAuthService{startErr: gh.ErrOrganizationAccessDenied})
	req := httptest.NewRequest(http.MethodPost, "/v1/github/connect/start", bytes.NewBufferString(`{"organizationId":"`+uuid.New().String()+`"}`))
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}
