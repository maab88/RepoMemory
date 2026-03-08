package dto

import (
	"time"

	"github.com/google/uuid"
	servicememory "github.com/maab88/repomemory/apps/api/internal/services/memory"
)

type MemoryEntryDTO struct {
	ID             uuid.UUID `json:"id"`
	RepositoryID   uuid.UUID `json:"repositoryId"`
	OrganizationID uuid.UUID `json:"organizationId"`
	Type           string    `json:"type"`
	Title          string    `json:"title"`
	Summary        string    `json:"summary"`
	WhyItMatters   string    `json:"whyItMatters,omitempty"`
	ImpactedAreas  []string  `json:"impactedAreas"`
	Risks          []string  `json:"risks"`
	FollowUps      []string  `json:"followUps"`
	GeneratedBy    string    `json:"generatedBy"`
	SourceURL      string    `json:"sourceUrl,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type MemorySourceDTO struct {
	SourceType   string `json:"sourceType"`
	SourceURL    string `json:"sourceUrl,omitempty"`
	DisplayLabel string `json:"displayLabel"`
}

type MemoryEntryDetailDTO struct {
	MemoryEntryDTO
	Sources []MemorySourceDTO `json:"sources"`
}

func ToMemoryEntryDTO(entry servicememory.MemoryEntry) MemoryEntryDTO {
	return MemoryEntryDTO{
		ID:             entry.ID,
		RepositoryID:   entry.RepositoryID,
		OrganizationID: entry.OrganizationID,
		Type:           entry.Type,
		Title:          entry.Title,
		Summary:        entry.Summary,
		WhyItMatters:   entry.WhyItMatters,
		ImpactedAreas:  entry.ImpactedAreas,
		Risks:          entry.Risks,
		FollowUps:      entry.FollowUps,
		GeneratedBy:    entry.GeneratedBy,
		SourceURL:      entry.SourceURL,
		CreatedAt:      entry.CreatedAt,
	}
}

func ToMemoryEntryDetailDTO(entry servicememory.MemoryEntry) MemoryEntryDetailDTO {
	sources := make([]MemorySourceDTO, 0, len(entry.Sources))
	for _, source := range entry.Sources {
		sources = append(sources, MemorySourceDTO{
			SourceType:   source.SourceType,
			SourceURL:    source.SourceURL,
			DisplayLabel: source.DisplayLabel,
		})
	}
	return MemoryEntryDetailDTO{
		MemoryEntryDTO: ToMemoryEntryDTO(entry),
		Sources:        sources,
	}
}
