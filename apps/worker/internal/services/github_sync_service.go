package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

const (
	initialSyncPerPage = 100
)

type SyncStore interface {
	GetRepositoryForSync(ctx context.Context, repositoryID uuid.UUID) (jobs.RepositoryForSync, error)
	GetGitHubAccessTokenForUser(ctx context.Context, userID uuid.UUID) (string, error)
	UpsertPullRequest(ctx context.Context, record jobs.PullRequestSyncRecord) error
	UpsertIssue(ctx context.Context, record jobs.IssueSyncRecord) error
	SetRepositoryLastPRSyncAt(ctx context.Context, repositoryID uuid.UUID, at time.Time) error
	SetRepositoryLastIssueSyncAt(ctx context.Context, repositoryID uuid.UUID, at time.Time) error
}

type GitHubSyncClient interface {
	ListPullRequests(ctx context.Context, accessToken, ownerLogin, repoName string) ([]GitHubPullRequest, error)
	ListIssues(ctx context.Context, accessToken, ownerLogin, repoName string) ([]GitHubIssue, error)
}

type GitHubSyncService struct {
	store  SyncStore
	client GitHubSyncClient
}

func NewGitHubSyncService(store SyncStore, client GitHubSyncClient) *GitHubSyncService {
	return &GitHubSyncService{store: store, client: client}
}

func (s *GitHubSyncService) Run(ctx context.Context, payload jobs.RepoInitialSyncPayload) error {
	// v1 initial sync fetches the latest single page (up to 100) PRs and issues sorted by updated_at desc.
	// PR sync and issue sync are intentionally separate: PRs are persisted first, then issues.
	// If issues fail after PR success, PRs remain synced and the caller records a partial-failure status/error.
	repo, err := s.store.GetRepositoryForSync(ctx, payload.RepositoryID)
	if err != nil {
		return fmt.Errorf("load repository for sync: %w", err)
	}

	accessToken, err := s.store.GetGitHubAccessTokenForUser(ctx, payload.TriggeredByUserID)
	if err != nil {
		return fmt.Errorf("load github token: %w", err)
	}

	prs, err := s.client.ListPullRequests(ctx, accessToken, repo.OwnerLogin, repo.Name)
	if err != nil {
		return fmt.Errorf("sync pull requests: %w", err)
	}
	prSyncTime := time.Now().UTC()
	for _, pr := range prs {
		record := MapPullRequestToSyncRecord(repo.ID, pr, prSyncTime)
		if err := s.store.UpsertPullRequest(ctx, record); err != nil {
			return fmt.Errorf("upsert pull request %d: %w", pr.ID, err)
		}
	}
	if err := s.store.SetRepositoryLastPRSyncAt(ctx, repo.ID, prSyncTime); err != nil {
		return fmt.Errorf("update pull request sync timestamp: %w", err)
	}

	issues, err := s.client.ListIssues(ctx, accessToken, repo.OwnerLogin, repo.Name)
	if err != nil {
		return fmt.Errorf("sync issues after pull requests synced: %w", err)
	}
	issueSyncTime := time.Now().UTC()
	for _, issue := range issues {
		record, ok := MapIssueToSyncRecord(repo.ID, issue, issueSyncTime)
		if !ok {
			continue
		}
		if err := s.store.UpsertIssue(ctx, record); err != nil {
			return fmt.Errorf("upsert issue %d: %w", issue.ID, err)
		}
	}
	if err := s.store.SetRepositoryLastIssueSyncAt(ctx, repo.ID, issueSyncTime); err != nil {
		return fmt.Errorf("update issue sync timestamp: %w", err)
	}

	return nil
}

type HTTPGitHubSyncClient struct {
	client     *http.Client
	apiBaseURL string
}

func NewHTTPGitHubSyncClient(apiBaseURL string) *HTTPGitHubSyncClient {
	if apiBaseURL == "" {
		apiBaseURL = "https://api.github.com"
	}
	return &HTTPGitHubSyncClient{
		client:     &http.Client{Timeout: 15 * time.Second},
		apiBaseURL: strings.TrimRight(apiBaseURL, "/"),
	}
}

func (c *HTTPGitHubSyncClient) ListPullRequests(ctx context.Context, accessToken, ownerLogin, repoName string) ([]GitHubPullRequest, error) {
	params := url.Values{}
	params.Set("state", "all")
	params.Set("sort", "updated")
	params.Set("direction", "desc")
	params.Set("per_page", fmt.Sprintf("%d", initialSyncPerPage))

	url := fmt.Sprintf("%s/repos/%s/%s/pulls?%s", c.apiBaseURL, ownerLogin, repoName, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	setGitHubHeaders(req, accessToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github pulls endpoint returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var prs []GitHubPullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, err
	}
	return prs, nil
}

func (c *HTTPGitHubSyncClient) ListIssues(ctx context.Context, accessToken, ownerLogin, repoName string) ([]GitHubIssue, error) {
	params := url.Values{}
	params.Set("state", "all")
	params.Set("sort", "updated")
	params.Set("direction", "desc")
	params.Set("per_page", fmt.Sprintf("%d", initialSyncPerPage))

	url := fmt.Sprintf("%s/repos/%s/%s/issues?%s", c.apiBaseURL, ownerLogin, repoName, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	setGitHubHeaders(req, accessToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github issues endpoint returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var issues []GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}
	return issues, nil
}

func setGitHubHeaders(req *http.Request, accessToken string) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "RepoMemory-Worker")
}

var _ GitHubSyncClient = (*HTTPGitHubSyncClient)(nil)
