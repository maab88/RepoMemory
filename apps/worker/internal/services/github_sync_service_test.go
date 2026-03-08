package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type fakeSyncStore struct {
	repo               jobs.RepositoryForSync
	token              string
	prUpserts          map[int64]jobs.PullRequestSyncRecord
	issueUpserts       map[int64]jobs.IssueSyncRecord
	lastPRSyncAtSet    bool
	lastIssueSyncAtSet bool
}

func (f *fakeSyncStore) GetRepositoryForSync(context.Context, uuid.UUID) (jobs.RepositoryForSync, error) {
	return f.repo, nil
}
func (f *fakeSyncStore) GetGitHubAccessTokenForUser(context.Context, uuid.UUID) (string, error) {
	return f.token, nil
}
func (f *fakeSyncStore) UpsertPullRequest(_ context.Context, record jobs.PullRequestSyncRecord) error {
	if f.prUpserts == nil {
		f.prUpserts = map[int64]jobs.PullRequestSyncRecord{}
	}
	f.prUpserts[record.GitHubPrID] = record
	return nil
}
func (f *fakeSyncStore) UpsertIssue(_ context.Context, record jobs.IssueSyncRecord) error {
	if f.issueUpserts == nil {
		f.issueUpserts = map[int64]jobs.IssueSyncRecord{}
	}
	f.issueUpserts[record.GitHubIssueID] = record
	return nil
}
func (f *fakeSyncStore) SetRepositoryLastPRSyncAt(context.Context, uuid.UUID, time.Time) error {
	f.lastPRSyncAtSet = true
	return nil
}
func (f *fakeSyncStore) SetRepositoryLastIssueSyncAt(context.Context, uuid.UUID, time.Time) error {
	f.lastIssueSyncAtSet = true
	return nil
}

type fakeGitHubSyncClient struct {
	prs      []GitHubPullRequest
	issues   []GitHubIssue
	prErr    error
	issueErr error
}

func (f *fakeGitHubSyncClient) ListPullRequests(context.Context, string, string, string) ([]GitHubPullRequest, error) {
	if f.prErr != nil {
		return nil, f.prErr
	}
	return f.prs, nil
}
func (f *fakeGitHubSyncClient) ListIssues(context.Context, string, string, string) ([]GitHubIssue, error) {
	if f.issueErr != nil {
		return nil, f.issueErr
	}
	return f.issues, nil
}

func TestGitHubSyncServiceSuccessAndIdempotentUpsert(t *testing.T) {
	repositoryID := uuid.New()
	store := &fakeSyncStore{
		repo: jobs.RepositoryForSync{
			ID:         repositoryID,
			OwnerLogin: "octo",
			Name:       "repo",
		},
		token: "token",
	}
	client := &fakeGitHubSyncClient{
		prs: []GitHubPullRequest{{ID: 1, Number: 1, Title: "pr", State: "open", HTMLURL: "https://github.com/octo/repo/pull/1", CreatedAt: time.Now().Add(-time.Hour), UpdatedAt: time.Now()}},
		issues: []GitHubIssue{
			{ID: 2, Number: 2, Title: "issue", State: "open", HTMLURL: "https://github.com/octo/repo/issues/2", CreatedAt: time.Now().Add(-time.Hour), UpdatedAt: time.Now()},
			{ID: 3, Number: 3, Title: "pr-issue-shape", State: "open", HTMLURL: "https://github.com/octo/repo/issues/3", PullRequest: &struct {
				URL string `json:"url"`
			}{URL: "https://api.github.com/repos/octo/repo/pulls/3"}, CreatedAt: time.Now().Add(-time.Hour), UpdatedAt: time.Now()},
		},
	}
	service := NewGitHubSyncService(store, client)
	payload := jobs.RepoInitialSyncPayload{RepositoryID: repositoryID, TriggeredByUserID: uuid.New()}

	if err := service.Run(context.Background(), payload); err != nil {
		t.Fatalf("first sync run failed: %v", err)
	}
	if err := service.Run(context.Background(), payload); err != nil {
		t.Fatalf("second sync run failed: %v", err)
	}

	if len(store.prUpserts) != 1 {
		t.Fatalf("expected 1 unique pr upsert, got %d", len(store.prUpserts))
	}
	if len(store.issueUpserts) != 1 {
		t.Fatalf("expected 1 unique issue upsert, got %d", len(store.issueUpserts))
	}
	if !store.lastPRSyncAtSet || !store.lastIssueSyncAtSet {
		t.Fatalf("expected both sync timestamps to be set")
	}
}

func TestGitHubSyncServicePartialFailureAfterPRSync(t *testing.T) {
	repositoryID := uuid.New()
	store := &fakeSyncStore{
		repo: jobs.RepositoryForSync{
			ID:         repositoryID,
			OwnerLogin: "octo",
			Name:       "repo",
		},
		token: "token",
	}
	client := &fakeGitHubSyncClient{
		prs:      []GitHubPullRequest{{ID: 1, Number: 1, Title: "pr", State: "open", HTMLURL: "https://github.com/octo/repo/pull/1", CreatedAt: time.Now().Add(-time.Hour), UpdatedAt: time.Now()}},
		issueErr: errors.New("github issues failed"),
	}
	service := NewGitHubSyncService(store, client)
	payload := jobs.RepoInitialSyncPayload{RepositoryID: repositoryID, TriggeredByUserID: uuid.New()}

	err := service.Run(context.Background(), payload)
	if err == nil {
		t.Fatal("expected partial failure error")
	}
	if !strings.Contains(err.Error(), "after pull requests synced") {
		t.Fatalf("expected partial failure message, got %v", err)
	}
	if !store.lastPRSyncAtSet {
		t.Fatal("expected pr sync timestamp to be set")
	}
	if store.lastIssueSyncAtSet {
		t.Fatal("did not expect issue sync timestamp to be set")
	}
}
