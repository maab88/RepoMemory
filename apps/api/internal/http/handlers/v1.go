package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/auth"
	gh "github.com/maab88/repomemory/apps/api/internal/github"
	"github.com/maab88/repomemory/apps/api/internal/http/response"
	"github.com/maab88/repomemory/apps/api/internal/org"
	servicejobs "github.com/maab88/repomemory/apps/api/internal/services/jobs"
	servicememory "github.com/maab88/repomemory/apps/api/internal/services/memory"
	servicerepositories "github.com/maab88/repomemory/apps/api/internal/services/repositories"
	servicesearch "github.com/maab88/repomemory/apps/api/internal/services/search"
)

type OrganizationService interface {
	CreateOrganization(ctx context.Context, userID uuid.UUID, name string) (org.OrganizationWithRole, error)
	ListOrganizations(ctx context.Context, userID uuid.UUID) ([]org.OrganizationWithRole, error)
	GetOrganization(ctx context.Context, userID, orgID uuid.UUID) (org.OrganizationWithRole, error)
}

type V1Handler struct {
	orgService        OrganizationService
	githubService     GitHubService
	jobService        JobService
	repositoryService RepositoryService
	memoryService     MemoryService
	searchService     SearchService
}

func NewV1Handler(
	orgService OrganizationService,
	githubService GitHubService,
	jobService JobService,
	repositoryService RepositoryService,
	memoryService MemoryService,
	searchService SearchService,
) *V1Handler {
	return &V1Handler{
		orgService:        orgService,
		githubService:     githubService,
		jobService:        jobService,
		repositoryService: repositoryService,
		memoryService:     memoryService,
		searchService:     searchService,
	}
}

type GitHubService interface {
	StartConnect(ctx context.Context, input gh.OAuthStartInput) (string, error)
	HandleCallback(ctx context.Context, input gh.OAuthCallbackInput) (gh.GitHubConnectionSummary, error)
	ListGitHubRepositories(ctx context.Context, userID uuid.UUID) ([]gh.GitHubRepository, error)
	ImportRepositories(ctx context.Context, input gh.ImportRepositoriesInput) ([]gh.ImportedRepository, error)
}

type JobService interface {
	GetJob(ctx context.Context, userID, jobID uuid.UUID) (servicejobs.Job, error)
}

type RepositoryService interface {
	ListRepositoriesForUser(ctx context.Context, userID uuid.UUID) ([]servicerepositories.Repository, error)
	ListOrganizationRepositories(ctx context.Context, userID, organizationID uuid.UUID) ([]servicerepositories.Repository, error)
	GetRepository(ctx context.Context, userID, repositoryID uuid.UUID) (servicerepositories.Repository, error)
	ListDigests(ctx context.Context, userID, repositoryID uuid.UUID) ([]servicerepositories.Digest, error)
	TriggerInitialSync(ctx context.Context, userID, repositoryID uuid.UUID) (servicejobs.Job, error)
	TriggerMemoryGeneration(ctx context.Context, userID, repositoryID uuid.UUID) (servicejobs.Job, error)
	TriggerDigestGeneration(ctx context.Context, userID, repositoryID uuid.UUID) (servicejobs.Job, error)
}

type MemoryService interface {
	ListRepositoryMemory(ctx context.Context, userID, repositoryID uuid.UUID) ([]servicememory.MemoryEntry, error)
	GetRepositoryMemoryEntry(ctx context.Context, userID, repositoryID, memoryID uuid.UUID) (servicememory.MemoryEntry, error)
}

type SearchService interface {
	SearchMemory(ctx context.Context, input servicesearch.MemorySearchInput) (servicesearch.MemorySearchResponse, error)
}

type meResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email,omitempty"`
	DisplayName string    `json:"displayName"`
	AvatarURL   string    `json:"avatarUrl,omitempty"`
}

type organizationResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
	Role string    `json:"role"`
}

func toOrganizationResponse(value org.OrganizationWithRole) organizationResponse {
	return organizationResponse{
		ID:   value.ID,
		Name: value.Name,
		Slug: value.Slug,
		Role: value.Role,
	}
}

func toOrganizationResponseList(value []org.OrganizationWithRole) []organizationResponse {
	out := make([]organizationResponse, 0, len(value))
	for _, item := range value {
		out = append(out, toOrganizationResponse(item))
	}
	return out
}

func (h *V1Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}

	response.WriteData(w, http.StatusOK, meResponse{
		ID:          currentUser.ID,
		Email:       currentUser.Email,
		DisplayName: currentUser.DisplayName,
		AvatarURL:   currentUser.AvatarURL,
	})
}

func (h *V1Handler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}

	orgs, err := h.orgService.ListOrganizations(r.Context(), currentUser.ID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list organizations")
		return
	}

	response.WriteData(w, http.StatusOK, toOrganizationResponseList(orgs))
}

type createOrganizationRequest struct {
	Name string `json:"name"`
}

func (h *V1Handler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}

	var req createOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid request payload")
		return
	}

	created, err := h.orgService.CreateOrganization(r.Context(), currentUser.ID, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, org.ErrInvalidOrganizationName):
			response.WriteError(w, http.StatusBadRequest, "validation_error", "organization name must be 2-80 characters")
		case errors.Is(err, org.ErrOrganizationConflict):
			response.WriteError(w, http.StatusConflict, "conflict", "organization slug already exists")
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to create organization")
		}
		return
	}

	response.WriteData(w, http.StatusCreated, toOrganizationResponse(created))
}

func (h *V1Handler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}

	orgIDRaw := chi.URLParam(r, "orgId")
	orgID, err := uuid.Parse(orgIDRaw)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid organization id")
		return
	}

	orgData, err := h.orgService.GetOrganization(r.Context(), currentUser.ID, orgID)
	if err != nil {
		switch {
		case errors.Is(err, org.ErrOrganizationForbidden):
			response.WriteError(w, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, org.ErrOrganizationNotFound):
			response.WriteError(w, http.StatusNotFound, "not_found", "organization not found")
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to fetch organization")
		}
		return
	}

	response.WriteData(w, http.StatusOK, toOrganizationResponse(orgData))
}

type startGitHubConnectRequest struct {
	OrganizationID string `json:"organizationId,omitempty"`
}

type startGitHubConnectResponse struct {
	RedirectURL string `json:"redirectUrl"`
}

type githubCallbackResponse struct {
	Connected bool                       `json:"connected"`
	Account   gh.GitHubConnectionSummary `json:"account"`
}

type githubRepositoryDTO struct {
	GitHubRepoID  string `json:"githubRepoId"`
	OwnerLogin    string `json:"ownerLogin"`
	Name          string `json:"name"`
	FullName      string `json:"fullName"`
	Private       bool   `json:"private"`
	DefaultBranch string `json:"defaultBranch"`
	HTMLURL       string `json:"htmlUrl"`
	Description   string `json:"description,omitempty"`
}

type githubRepositoriesListResponse struct {
	Repositories []githubRepositoryDTO `json:"repositories"`
}

type importGitHubRepositoriesRequest struct {
	OrganizationID string                `json:"organizationId"`
	Repositories   []githubRepositoryDTO `json:"repositories"`
}

type importGitHubRepositoriesResponse struct {
	ImportedRepositories []gh.ImportedRepository `json:"importedRepositories"`
}

func (h *V1Handler) StartGitHubConnect(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}

	var req startGitHubConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid request payload")
		return
	}

	var organizationID *uuid.UUID
	if req.OrganizationID != "" {
		parsed, err := uuid.Parse(req.OrganizationID)
		if err != nil {
			response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid organization id")
			return
		}
		organizationID = &parsed
	}

	redirectURL, err := h.githubService.StartConnect(r.Context(), gh.OAuthStartInput{
		UserID:         currentUser.ID,
		OrganizationID: organizationID,
	})
	if err != nil {
		switch {
		case errors.Is(err, gh.ErrOrganizationAccessDenied):
			response.WriteError(w, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, gh.ErrOAuthNotConfigured):
			response.WriteError(w, http.StatusServiceUnavailable, "github_oauth_not_configured", "github oauth is not configured")
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to start github connect")
		}
		return
	}

	response.WriteData(w, http.StatusOK, startGitHubConnectResponse{RedirectURL: redirectURL})
}

func (h *V1Handler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	account, err := h.githubService.HandleCallback(r.Context(), gh.OAuthCallbackInput{
		UserID: currentUser.ID,
		Code:   code,
		State:  state,
	})
	if err != nil {
		switch {
		case errors.Is(err, gh.ErrOAuthCodeMissing), errors.Is(err, gh.ErrOAuthStateMissing), errors.Is(err, gh.ErrInvalidState), errors.Is(err, gh.ErrStateExpired):
			response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid github oauth callback request")
		case errors.Is(err, gh.ErrStateUserMismatch), errors.Is(err, gh.ErrOrganizationAccessDenied):
			response.WriteError(w, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, gh.ErrTokenExchangeFailed):
			response.WriteError(w, http.StatusBadGateway, "github_oauth_exchange_failed", "github oauth token exchange failed")
		case errors.Is(err, gh.ErrGitHubUserFetchFailed):
			response.WriteError(w, http.StatusBadGateway, "github_user_fetch_failed", "failed to fetch github user")
		case errors.Is(err, gh.ErrOAuthNotConfigured):
			response.WriteError(w, http.StatusServiceUnavailable, "github_oauth_not_configured", "github oauth is not configured")
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to complete github oauth callback")
		}
		return
	}

	response.WriteData(w, http.StatusOK, githubCallbackResponse{Connected: true, Account: account})
}

func (h *V1Handler) ListGitHubRepositories(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}

	repos, err := h.githubService.ListGitHubRepositories(r.Context(), currentUser.ID)
	if err != nil {
		switch {
		case errors.Is(err, gh.ErrGitHubNotConnected):
			response.WriteError(w, http.StatusBadRequest, "github_not_connected", "github account is not connected")
		default:
			response.WriteError(w, http.StatusBadGateway, "github_repositories_fetch_failed", "failed to fetch github repositories")
		}
		return
	}

	payload := make([]githubRepositoryDTO, 0, len(repos))
	for _, repo := range repos {
		payload = append(payload, githubRepositoryDTO{
			GitHubRepoID:  strconv.FormatInt(repo.GitHubRepoID, 10),
			OwnerLogin:    repo.OwnerLogin,
			Name:          repo.Name,
			FullName:      repo.FullName,
			Private:       repo.Private,
			DefaultBranch: repo.DefaultBranch,
			HTMLURL:       repo.HTMLURL,
			Description:   repo.Description,
		})
	}

	response.WriteData(w, http.StatusOK, githubRepositoriesListResponse{Repositories: payload})
}

func (h *V1Handler) ImportGitHubRepositories(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}

	var req importGitHubRepositoriesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid request payload")
		return
	}

	organizationID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid organization id")
		return
	}

	repos := make([]gh.GitHubRepository, 0, len(req.Repositories))
	for _, repo := range req.Repositories {
		githubRepoID, parseErr := strconv.ParseInt(repo.GitHubRepoID, 10, 64)
		if parseErr != nil {
			response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid github repository id")
			return
		}
		repos = append(repos, gh.GitHubRepository{
			GitHubRepoID:  githubRepoID,
			OwnerLogin:    repo.OwnerLogin,
			Name:          repo.Name,
			FullName:      repo.FullName,
			Private:       repo.Private,
			DefaultBranch: repo.DefaultBranch,
			HTMLURL:       repo.HTMLURL,
			Description:   repo.Description,
		})
	}

	imported, err := h.githubService.ImportRepositories(r.Context(), gh.ImportRepositoriesInput{
		UserID:         currentUser.ID,
		OrganizationID: organizationID,
		Repositories:   repos,
	})
	if err != nil {
		switch {
		case errors.Is(err, gh.ErrOrganizationAccessDenied):
			response.WriteError(w, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, gh.ErrGitHubNotConnected):
			response.WriteError(w, http.StatusBadRequest, "github_not_connected", "github account is not connected")
		case errors.Is(err, gh.ErrImportRepositoriesEmpty), errors.Is(err, gh.ErrInvalidRepositoryPayload):
			response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid repositories payload")
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to import repositories")
		}
		return
	}

	response.WriteData(w, http.StatusOK, importGitHubRepositoriesResponse{ImportedRepositories: imported})
}
