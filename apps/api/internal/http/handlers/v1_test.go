package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/auth"
	gh "github.com/maab88/repomemory/apps/api/internal/github"
	"github.com/maab88/repomemory/apps/api/internal/org"
	servicejobs "github.com/maab88/repomemory/apps/api/internal/services/jobs"
	servicerepositories "github.com/maab88/repomemory/apps/api/internal/services/repositories"
)

type fakeResolver struct{}

func (f fakeResolver) Resolve(_ context.Context, input auth.MockUserInput) (auth.CurrentUser, error) {
	return auth.CurrentUser{
		ID:          uuid.NewSHA1(uuid.NameSpaceOID, []byte(input.RawID)),
		DisplayName: "Dev Tester",
		Email:       input.Email,
		CreatedAt:   time.Date(2026, 3, 9, 12, 0, 0, 0, time.UTC),
	}, nil
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
	redirectURL  string
	account      gh.GitHubConnectionSummary
	repositories []gh.GitHubRepository
	imported     []gh.ImportedRepository
	startErr     error
	callbackErr  error
	listErr      error
	importErr    error
	callbacks    int
}

type fakeJobService struct {
	job servicejobs.Job
	err error
}

func (f *fakeJobService) GetJob(_ context.Context, _, _ uuid.UUID) (servicejobs.Job, error) {
	return f.job, f.err
}

type fakeRepositoryService struct {
	listForUser  []servicerepositories.Repository
	repositories []servicerepositories.Repository
	repository   servicerepositories.Repository
	digests      []servicerepositories.Digest
	job          servicejobs.Job
	listErr      error
	getErr       error
	syncErr      error
	memoryErr    error
	digestErr    error
}

func (f *fakeRepositoryService) ListRepositoriesForUser(_ context.Context, _ uuid.UUID) ([]servicerepositories.Repository, error) {
	return f.listForUser, nil
}

func (f *fakeRepositoryService) ListOrganizationRepositories(_ context.Context, _, _ uuid.UUID) ([]servicerepositories.Repository, error) {
	return f.repositories, f.listErr
}

func (f *fakeRepositoryService) GetRepository(_ context.Context, _, _ uuid.UUID) (servicerepositories.Repository, error) {
	return f.repository, f.getErr
}
func (f *fakeRepositoryService) ListDigests(_ context.Context, _, _ uuid.UUID) ([]servicerepositories.Digest, error) {
	return f.digests, f.digestErr
}

func (f *fakeRepositoryService) TriggerInitialSync(_ context.Context, _, _ uuid.UUID) (servicejobs.Job, error) {
	return f.job, f.syncErr
}

func (f *fakeRepositoryService) TriggerMemoryGeneration(_ context.Context, _, _ uuid.UUID) (servicejobs.Job, error) {
	return f.job, f.memoryErr
}
func (f *fakeRepositoryService) TriggerDigestGeneration(_ context.Context, _, _ uuid.UUID) (servicejobs.Job, error) {
	return f.job, f.digestErr
}

func (f *fakeGitHubOAuthService) StartConnect(_ context.Context, _ gh.OAuthStartInput) (string, error) {
	return f.redirectURL, f.startErr
}

func (f *fakeGitHubOAuthService) HandleCallback(_ context.Context, _ gh.OAuthCallbackInput) (gh.GitHubConnectionSummary, error) {
	f.callbacks++
	return f.account, f.callbackErr
}

func (f *fakeGitHubOAuthService) ListGitHubRepositories(_ context.Context, _ uuid.UUID) ([]gh.GitHubRepository, error) {
	return f.repositories, f.listErr
}

func (f *fakeGitHubOAuthService) ImportRepositories(_ context.Context, _ gh.ImportRepositoriesInput) ([]gh.ImportedRepository, error) {
	return f.imported, f.importErr
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

func newTestRouterWithGitHub(orgSvc OrganizationService, githubSvc GitHubService) http.Handler {
	h := NewV1Handler(orgSvc, githubSvc, &fakeJobService{}, &fakeRepositoryService{}, &noopMemoryService{}, &noopSearchService{})
	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/me", h.GetMe)
		r.Get("/organizations", h.ListOrganizations)
		r.Post("/organizations", h.CreateOrganization)
		r.Get("/organizations/{orgId}", h.GetOrganization)
		r.Post("/github/connect/start", h.StartGitHubConnect)
		r.Get("/github/callback", h.GitHubCallback)
		r.Get("/github/repositories", h.ListGitHubRepositories)
		r.Post("/github/repositories/import", h.ImportGitHubRepositories)
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

	var payload struct {
		Data struct {
			ID          string `json:"id"`
			Email       string `json:"email"`
			DisplayName string `json:"displayName"`
			CreatedAt   string `json:"createdAt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Data.Email != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %s", payload.Data.Email)
	}
	if payload.Data.DisplayName != "Dev Tester" {
		t.Fatalf("expected displayName Dev Tester, got %s", payload.Data.DisplayName)
	}
	if payload.Data.CreatedAt == "" {
		t.Fatal("expected createdAt in /v1/me response")
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
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Error.Code != "OAUTH_CALLBACK_FAILED" {
		t.Fatalf("expected OAUTH_CALLBACK_FAILED, got %s", payload.Error.Code)
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

func TestListGitHubRepositoriesSuccess(t *testing.T) {
	router := newTestRouterWithGitHub(&fakeOrgService{}, &fakeGitHubOAuthService{
		repositories: []gh.GitHubRepository{
			{
				GitHubRepoID:  123,
				OwnerLogin:    "octocat",
				Name:          "repo-memory",
				FullName:      "octocat/repo-memory",
				Private:       true,
				DefaultBranch: "main",
				HTMLURL:       "https://github.com/octocat/repo-memory",
				Description:   "Internal tools",
			},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/github/repositories", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestListGitHubRepositoriesNotConnected(t *testing.T) {
	router := newTestRouterWithGitHub(&fakeOrgService{}, &fakeGitHubOAuthService{listErr: gh.ErrGitHubNotConnected})
	req := httptest.NewRequest(http.MethodGet, "/v1/github/repositories", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Error.Code != "GITHUB_RECONNECT_REQUIRED" {
		t.Fatalf("expected GITHUB_RECONNECT_REQUIRED, got %s", payload.Error.Code)
	}
}

func TestListGitHubRepositoriesRateLimited(t *testing.T) {
	router := newTestRouterWithGitHub(&fakeOrgService{}, &fakeGitHubOAuthService{listErr: gh.ErrGitHubRateLimited})
	req := httptest.NewRequest(http.MethodGet, "/v1/github/repositories", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Error.Code != "GITHUB_RATE_LIMITED" {
		t.Fatalf("expected GITHUB_RATE_LIMITED, got %s", payload.Error.Code)
	}
}

func TestImportGitHubRepositoriesForbidden(t *testing.T) {
	router := newTestRouterWithGitHub(&fakeOrgService{}, &fakeGitHubOAuthService{importErr: gh.ErrOrganizationAccessDenied})
	req := httptest.NewRequest(http.MethodPost, "/v1/github/repositories/import", bytes.NewBufferString(`{"organizationId":"`+uuid.New().String()+`","repositories":[{"githubRepoId":"123","ownerLogin":"octocat","name":"repo","fullName":"octocat/repo","private":true,"defaultBranch":"main","htmlUrl":"https://github.com/octocat/repo"}]}`))
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestImportGitHubRepositoriesSuccess(t *testing.T) {
	imported := []gh.ImportedRepository{
		{
			ID:             uuid.New(),
			OrganizationID: uuid.New(),
			GitHubRepoID:   "123",
			OwnerLogin:     "octocat",
			Name:           "repo-memory",
			FullName:       "octocat/repo-memory",
			Private:        true,
			DefaultBranch:  "main",
			HTMLURL:        "https://github.com/octocat/repo-memory",
		},
	}
	router := newTestRouterWithGitHub(&fakeOrgService{}, &fakeGitHubOAuthService{imported: imported})
	req := httptest.NewRequest(http.MethodPost, "/v1/github/repositories/import", bytes.NewBufferString(`{"organizationId":"`+imported[0].OrganizationID.String()+`","repositories":[{"githubRepoId":"123","ownerLogin":"octocat","name":"repo-memory","fullName":"octocat/repo-memory","private":true,"defaultBranch":"main","htmlUrl":"https://github.com/octocat/repo-memory"}]}`))
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestImportGitHubRepositoriesNotConnected(t *testing.T) {
	router := newTestRouterWithGitHub(&fakeOrgService{}, &fakeGitHubOAuthService{importErr: gh.ErrGitHubNotConnected})
	req := httptest.NewRequest(http.MethodPost, "/v1/github/repositories/import", bytes.NewBufferString(`{"organizationId":"`+uuid.New().String()+`","repositories":[{"githubRepoId":"123","ownerLogin":"octocat","name":"repo-memory","fullName":"octocat/repo-memory","private":true,"defaultBranch":"main","htmlUrl":"https://github.com/octocat/repo-memory"}]}`))
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
