package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/worker/internal/jobs"
)

type GitHubPullRequest struct {
	ID        int64      `json:"id"`
	Number    int32      `json:"number"`
	Title     string     `json:"title"`
	Body      *string    `json:"body"`
	State     string     `json:"state"`
	HTMLURL   string     `json:"html_url"`
	MergedAt  *time.Time `json:"merged_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
}

type GitHubIssue struct {
	ID        int64      `json:"id"`
	Number    int32      `json:"number"`
	Title     string     `json:"title"`
	Body      *string    `json:"body"`
	State     string     `json:"state"`
	HTMLURL   string     `json:"html_url"`
	ClosedAt  *time.Time `json:"closed_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	User      struct {
		Login string `json:"login"`
	} `json:"user"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
	PullRequest *struct {
		URL string `json:"url"`
	} `json:"pull_request,omitempty"`
}

func MapPullRequestToSyncRecord(repositoryID uuid.UUID, pr GitHubPullRequest, syncedAt time.Time) jobs.PullRequestSyncRecord {
	body := ""
	if pr.Body != nil {
		body = *pr.Body
	}
	labels := make([]string, 0, len(pr.Labels))
	for _, label := range pr.Labels {
		if label.Name != "" {
			labels = append(labels, label.Name)
		}
	}

	return jobs.PullRequestSyncRecord{
		RepositoryID:      repositoryID,
		GitHubPrID:        pr.ID,
		GitHubPrNumber:    pr.Number,
		Title:             pr.Title,
		Body:              body,
		State:             pr.State,
		AuthorLogin:       pr.User.Login,
		HTMLURL:           pr.HTMLURL,
		MergedAt:          pr.MergedAt,
		ClosedAt:          pr.ClosedAt,
		Labels:            labels,
		CreatedAtExternal: pr.CreatedAt.UTC(),
		UpdatedAtExternal: pr.UpdatedAt.UTC(),
		SyncedAt:          syncedAt.UTC(),
	}
}

func MapIssueToSyncRecord(repositoryID uuid.UUID, issue GitHubIssue, syncedAt time.Time) (jobs.IssueSyncRecord, bool) {
	if issue.PullRequest != nil {
		return jobs.IssueSyncRecord{}, false
	}

	body := ""
	if issue.Body != nil {
		body = *issue.Body
	}
	labels := make([]string, 0, len(issue.Labels))
	for _, label := range issue.Labels {
		if label.Name != "" {
			labels = append(labels, label.Name)
		}
	}

	return jobs.IssueSyncRecord{
		RepositoryID:      repositoryID,
		GitHubIssueID:     issue.ID,
		GitHubIssueNumber: issue.Number,
		Title:             issue.Title,
		Body:              body,
		State:             issue.State,
		AuthorLogin:       issue.User.Login,
		HTMLURL:           issue.HTMLURL,
		ClosedAt:          issue.ClosedAt,
		Labels:            labels,
		CreatedAtExternal: issue.CreatedAt.UTC(),
		UpdatedAtExternal: issue.UpdatedAt.UTC(),
		SyncedAt:          syncedAt.UTC(),
	}, true
}
