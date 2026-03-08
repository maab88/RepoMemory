package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMapPullRequestToSyncRecord(t *testing.T) {
	repositoryID := uuid.New()
	now := time.Now().UTC()
	body := "details"
	pr := GitHubPullRequest{
		ID:        1001,
		Number:    42,
		Title:     "Add initial sync",
		Body:      &body,
		State:     "open",
		HTMLURL:   "https://github.com/octo/repo/pull/42",
		CreatedAt: now.Add(-2 * time.Hour),
		UpdatedAt: now,
	}
	pr.User.Login = "octocat"
	pr.Labels = []struct {
		Name string `json:"name"`
	}{{Name: "sync"}, {Name: "backend"}}

	record := MapPullRequestToSyncRecord(repositoryID, pr, now)
	if record.RepositoryID != repositoryID {
		t.Fatalf("repository mismatch")
	}
	if record.GitHubPrID != 1001 || record.GitHubPrNumber != 42 {
		t.Fatalf("unexpected pr ids: %+v", record)
	}
	if len(record.Labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(record.Labels))
	}
}

func TestMapIssueToSyncRecordFiltersPullRequestShapedIssues(t *testing.T) {
	repositoryID := uuid.New()
	now := time.Now().UTC()
	issue := GitHubIssue{
		ID:        2001,
		Number:    8,
		Title:     "Fix bug",
		State:     "open",
		HTMLURL:   "https://github.com/octo/repo/issues/8",
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now,
	}
	issue.User.Login = "hubot"

	record, ok := MapIssueToSyncRecord(repositoryID, issue, now)
	if !ok {
		t.Fatal("expected issue to map")
	}
	if record.GitHubIssueID != 2001 {
		t.Fatalf("unexpected issue id: %d", record.GitHubIssueID)
	}

	issue.PullRequest = &struct {
		URL string `json:"url"`
	}{URL: "https://api.github.com/repos/octo/repo/pulls/8"}
	if _, ok := MapIssueToSyncRecord(repositoryID, issue, now); ok {
		t.Fatal("expected pull-request-shaped issue to be filtered out")
	}
}
