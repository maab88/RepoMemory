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
	servicesearch "github.com/maab88/repomemory/apps/api/internal/services/search"
)

type fakeSearchService struct {
	result servicesearch.MemorySearchResponse
	err    error
}

func (f *fakeSearchService) SearchMemory(_ context.Context, _ servicesearch.MemorySearchInput) (servicesearch.MemorySearchResponse, error) {
	return f.result, f.err
}

func TestSearchMemoryReturnsResults(t *testing.T) {
	searchSvc := &fakeSearchService{
		result: servicesearch.MemorySearchResponse{
			Query:    "retry",
			Page:     1,
			PageSize: 20,
			Total:    1,
			Results: []servicesearch.MemorySearchResult{
				{
					ID:             uuid.New(),
					RepositoryID:   uuid.New(),
					RepositoryName: "repo-memory",
					Type:           "pr_summary",
					Title:          "Retry update",
					SummarySnippet: "Moved retry scheduling.",
					SourceURL:      "https://github.com/org/repo/pull/1",
					CreatedAt:      time.Now().UTC(),
				},
			},
		},
	}
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, &noopRepositoryService{}, &noopMemoryService{}, searchSvc)

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/memory/search", h.SearchMemory)
	})

	orgID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/memory/search?organizationId="+orgID.String()+"&q=retry&page=1&pageSize=20", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestSearchMemoryUnauthorizedOrgRejected(t *testing.T) {
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, &noopRepositoryService{}, &noopMemoryService{}, &fakeSearchService{err: servicesearch.ErrForbidden})

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/memory/search", h.SearchMemory)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/memory/search?organizationId="+uuid.New().String()+"&q=retry", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestSearchMemoryRequiresOrganizationID(t *testing.T) {
	h := NewV1Handler(&noopOrgService{}, &noopGitHubService{}, &fakeJobQueryService{}, &noopRepositoryService{}, &noopMemoryService{}, &fakeSearchService{})

	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Use(auth.RequireMockAuth(fakeResolver{}))
		r.Get("/memory/search", h.SearchMemory)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/memory/search?q=retry", nil)
	req.Header.Set("x-user-id", "user-1")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
