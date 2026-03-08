package dto

import (
	"time"

	"github.com/google/uuid"
	servicerepositories "github.com/maab88/repomemory/apps/api/internal/services/repositories"
)

type DigestDTO struct {
	ID           uuid.UUID `json:"id"`
	RepositoryID uuid.UUID `json:"repositoryId"`
	PeriodStart  time.Time `json:"periodStart"`
	PeriodEnd    time.Time `json:"periodEnd"`
	Title        string    `json:"title"`
	Summary      string    `json:"summary"`
	BodyMarkdown string    `json:"bodyMarkdown,omitempty"`
	GeneratedBy  string    `json:"generatedBy"`
	CreatedAt    time.Time `json:"createdAt"`
}

func ToDigestDTO(d servicerepositories.Digest) DigestDTO {
	return DigestDTO{
		ID:           d.ID,
		RepositoryID: d.RepositoryID,
		PeriodStart:  d.PeriodStart,
		PeriodEnd:    d.PeriodEnd,
		Title:        d.Title,
		Summary:      d.Summary,
		BodyMarkdown: d.BodyMarkdown,
		GeneratedBy:  d.GeneratedBy,
		CreatedAt:    d.CreatedAt,
	}
}
