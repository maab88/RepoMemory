package repositories

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type RepositorySummary struct {
	ID               uuid.UUID
	OrganizationID   uuid.UUID
	GitHubRepoID     string
	OwnerLogin       string
	Name             string
	FullName         string
	Private          bool
	DefaultBranch    string
	HTMLURL          string
	Description      string
	ImportedAt       time.Time
	LastSyncStatus   string
	LastSyncTime     *time.Time
	PullRequestCount int
	IssueCount       int
	MemoryEntryCount int
}

type RepositoryRepository struct {
	queries *db.Queries
}

func NewRepositoryRepository(queries *db.Queries) *RepositoryRepository {
	return &RepositoryRepository{queries: queries}
}

func (r *RepositoryRepository) ListSummariesForOrganization(ctx context.Context, organizationID uuid.UUID) ([]RepositorySummary, error) {
	rows, err := r.queries.ListRepositorySummariesForOrganization(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	result := make([]RepositorySummary, 0, len(rows))
	for _, row := range rows {
		var lastSyncTime *time.Time
		if row.LastSyncTime.Valid {
			t := row.LastSyncTime.Time.UTC()
			lastSyncTime = &t
		}
		result = append(result, RepositorySummary{
			ID:               row.ID,
			OrganizationID:   row.OrganizationID,
			GitHubRepoID:     strconv.FormatInt(row.GithubRepoID, 10),
			OwnerLogin:       row.OwnerLogin,
			Name:             row.Name,
			FullName:         row.FullName,
			Private:          row.Private,
			DefaultBranch:    row.DefaultBranch,
			HTMLURL:          row.HtmlUrl,
			Description:      row.Description.String,
			ImportedAt:       row.ImportedAt.Time.UTC(),
			LastSyncStatus:   row.LastSyncStatus.String,
			LastSyncTime:     lastSyncTime,
			PullRequestCount: int(row.PullRequestCount),
			IssueCount:       int(row.IssueCount),
			MemoryEntryCount: int(row.MemoryEntryCount),
		})
	}

	return result, nil
}

func (r *RepositoryRepository) ListSummariesForUser(ctx context.Context, userID uuid.UUID) ([]RepositorySummary, error) {
	rows, err := r.queries.ListRepositorySummariesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]RepositorySummary, 0, len(rows))
	for _, row := range rows {
		var lastSyncTime *time.Time
		if row.LastSyncTime.Valid {
			t := row.LastSyncTime.Time.UTC()
			lastSyncTime = &t
		}
		result = append(result, RepositorySummary{
			ID:               row.ID,
			OrganizationID:   row.OrganizationID,
			GitHubRepoID:     strconv.FormatInt(row.GithubRepoID, 10),
			OwnerLogin:       row.OwnerLogin,
			Name:             row.Name,
			FullName:         row.FullName,
			Private:          row.Private,
			DefaultBranch:    row.DefaultBranch,
			HTMLURL:          row.HtmlUrl,
			Description:      row.Description.String,
			ImportedAt:       row.ImportedAt.Time.UTC(),
			LastSyncStatus:   row.LastSyncStatus.String,
			LastSyncTime:     lastSyncTime,
			PullRequestCount: int(row.PullRequestCount),
			IssueCount:       int(row.IssueCount),
			MemoryEntryCount: int(row.MemoryEntryCount),
		})
	}

	return result, nil
}

func (r *RepositoryRepository) GetByID(ctx context.Context, repositoryID uuid.UUID) (db.Repository, error) {
	return r.queries.GetRepositoryByID(ctx, repositoryID)
}
