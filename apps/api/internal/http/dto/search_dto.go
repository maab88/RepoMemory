package dto

import (
	"time"

	"github.com/google/uuid"
	servicesearch "github.com/maab88/repomemory/apps/api/internal/services/search"
)

type MemorySearchResultDTO struct {
	ID             uuid.UUID `json:"id"`
	RepositoryID   uuid.UUID `json:"repositoryId"`
	RepositoryName string    `json:"repositoryName"`
	Type           string    `json:"type"`
	Title          string    `json:"title"`
	SummarySnippet string    `json:"summarySnippet"`
	SourceURL      string    `json:"sourceUrl,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type MemorySearchDataDTO struct {
	Query    string                  `json:"query"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"pageSize"`
	Total    int                     `json:"total"`
	Results  []MemorySearchResultDTO `json:"results"`
}

func ToMemorySearchDataDTO(value servicesearch.MemorySearchResponse) MemorySearchDataDTO {
	results := make([]MemorySearchResultDTO, 0, len(value.Results))
	for _, item := range value.Results {
		results = append(results, MemorySearchResultDTO{
			ID:             item.ID,
			RepositoryID:   item.RepositoryID,
			RepositoryName: item.RepositoryName,
			Type:           item.Type,
			Title:          item.Title,
			SummarySnippet: item.SummarySnippet,
			SourceURL:      item.SourceURL,
			CreatedAt:      item.CreatedAt,
		})
	}

	return MemorySearchDataDTO{
		Query:    value.Query,
		Page:     value.Page,
		PageSize: value.PageSize,
		Total:    value.Total,
		Results:  results,
	}
}
