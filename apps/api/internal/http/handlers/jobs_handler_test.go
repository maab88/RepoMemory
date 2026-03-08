package handlers

import (
	"context"
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
	servicememory "github.com/maab88/repomemory/apps/api/internal/services/memory"
	servicerepositories "github.com/maab88/repomemory/apps/api/internal/services/repositories"
	servicesearch "github.com/maab88/repomemory/apps/api/internal/services/search"
)

type noopOrgService struct{}

func (n *noopOrgService) CreateOrganization(context.Context, uuid.UUID, string) (org.OrganizationWithRole, error) {
	return org.OrganizationWithRole{}, nil
}
func (n *noopOrgService) ListOrganizations(context.Context, uuid.UUID) ([]org.OrganizationWithRole, error) {
	return nil, nil
}
func (n *noopOrgService) GetOrganization(context.Context, uuid.UUID, uuid.UUID) (org.OrganizationWithRole, error) {
	return org.OrganizationWithRole{}, nil
}

type noopGitHubService struct{}

func (n *noopGitHubService) StartConnect(context.Context, gh.OAuthStartInput) (string, error) {
	return "", nil
}
func (n *noopGitHubService) HandleCallback(context.Context, gh.OAuthCallbackInput) (gh.GitHubConnectionSummary, error) {
	return gh.GitHubConnectionSummary{}, nil
}
func (n *noopGitHubService) ListGitHubRepositories(context.Context, uuid.UUID) ([]gh.GitHubRepository, error) {
	return nil, nil
}
func (n *noopGitHubService) ImportRepositories(context.Context, gh.ImportRepositoriesInput) ([]gh.ImportedRepository, error) {
	return nil, nil
}

type noopRepositoryService struct{}

func (n *noopRepositoryService) ListRepositoriesForUser(context.Context, uuid.UUID) ([]servicerepositories.Repository, error) {
	return nil, nil
}

func (n *noopRepositoryService) ListOrganizationRepositories(context.Context, uuid.UUID, uuid.UUID) ([]servicerepositories.Repository, error) {
	return nil, nil
}
func (n *noopRepositoryService) GetRepository(context.Context, uuid.UUID, uuid.UUID) (servicerepositories.Repository, error) {
	return servicerepositories.Repository{}, nil
}
func (n *noopRepositoryService) TriggerInitialSync(context.Context, uuid.UUID, uuid.UUID) (servicejobs.Job, error) {
	return servicejobs.Job{}, nil
}
func (n *noopRepositoryService) TriggerMemoryGeneration(context.Context, uuid.UUID, uuid.UUID) (servicejobs.Job, error) {
	return servicejobs.Job{}, nil
}

type noopMemoryService struct{}

func (n *noopMemoryService) ListRepositoryMemory(context.Context, uuid.UUID, uuid.UUID) ([]servicememory.MemoryEntry, error) {
	return []servicememory.MemoryEntry{}, nil
}

func (n *noopMemoryService) GetRepositoryMemoryEntry(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (servicememory.MemoryEntry, error) {
	return servicememory.MemoryEntry{}, servicememory.ErrMemoryEntryNotFound
}

type noopSearchService struct{}

func (n *noopSearchService) SearchMemory(context.Context, servicesearch.MemorySearchInput) (servicesearch.MemorySearchResponse, error) {
	return servicesearch.MemorySearchResponse{}, nil
}

type fakeJobQueryService struct {
	job servicejobs.Job
	err error
}

func (f *fakeJobQueryService) GetJob(_ context.Context, _, _ uuid.UUID) (servicejobs.Job, error) {
	return f.job, f.err
}

func TestGetJobReturnsRecord(t *testing.T) {
	jobSvc := &fakeJobQueryService{
		job: servicejobs.Job{
			ID:        uuid.New(),
			JobType:   "repo.initial_sync",
			Status:    "queued",
			QueueName: "default",
			Attempts:  0,
			Payload:   []byte(`{"repositoryId":"` + uuid.New().String() + `"}`),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		},
	}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, jobSvc, &noopRepositoryService{}, &noopMemoryService{}, &noopSearchService{})

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/jobs/{jobId}", h.GetJob)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/jobs/"+jobSvc.job.ID.String(), nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetJobUnauthorizedRejected(t *testing.T) {
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, &noopRepositoryService{}, &noopMemoryService{}, &noopSearchService{})
	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/jobs/{jobId}", h.GetJob)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/jobs/"+uuid.New().String(), nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestGetJobForbidden(t *testing.T) {
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{err: servicejobs.ErrJobForbidden}, &noopRepositoryService{}, &noopMemoryService{}, &noopSearchService{})
	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/jobs/{jobId}", h.GetJob)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/jobs/"+uuid.New().String(), nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}
