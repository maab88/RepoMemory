package dto

import (
	"time"

	"github.com/google/uuid"
	servicerepositories "github.com/maab88/repomemory/apps/api/internal/services/repositories"
)

type RepositoryDTO struct {
	ID               uuid.UUID  `json:"id"`
	OrganizationID   uuid.UUID  `json:"organizationId"`
	GitHubRepoID     string     `json:"githubRepoId"`
	OwnerLogin       string     `json:"ownerLogin"`
	Name             string     `json:"name"`
	FullName         string     `json:"fullName"`
	Private          bool       `json:"private"`
	DefaultBranch    string     `json:"defaultBranch"`
	HTMLURL          string     `json:"htmlUrl"`
	Description      string     `json:"description,omitempty"`
	ImportedAt       time.Time  `json:"importedAt"`
	LastSyncStatus   string     `json:"lastSyncStatus,omitempty"`
	LastSyncTime     *time.Time `json:"lastSyncTime"`
	PullRequestCount int        `json:"pullRequestCount"`
	IssueCount       int        `json:"issueCount"`
	MemoryEntryCount int        `json:"memoryEntryCount"`
}

func ToRepositoryDTO(repo servicerepositories.Repository) RepositoryDTO {
	return RepositoryDTO{
		ID:               repo.ID,
		OrganizationID:   repo.OrganizationID,
		GitHubRepoID:     repo.GitHubRepoID,
		OwnerLogin:       repo.OwnerLogin,
		Name:             repo.Name,
		FullName:         repo.FullName,
		Private:          repo.Private,
		DefaultBranch:    repo.DefaultBranch,
		HTMLURL:          repo.HTMLURL,
		Description:      repo.Description,
		ImportedAt:       repo.ImportedAt,
		LastSyncStatus:   repo.LastSyncStatus,
		LastSyncTime:     repo.LastSyncTime,
		PullRequestCount: repo.PullRequestCount,
		IssueCount:       repo.IssueCount,
		MemoryEntryCount: repo.MemoryEntryCount,
	}
}
