package search

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/maab88/repomemory/apps/api/internal/db"
	"github.com/maab88/repomemory/apps/api/internal/repositories"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

var (
	ErrForbidden = errors.New("search access denied")
	ErrNotFound  = errors.New("repository not found")
)

type MembershipChecker interface {
	UserHasMembership(ctx context.Context, arg db.UserHasMembershipParams) (bool, error)
}

type RepositoryReader interface {
	GetByID(ctx context.Context, repositoryID uuid.UUID) (db.Repository, error)
}

type MemorySearcher interface {
	Search(ctx context.Context, input repositories.MemorySearchInput) ([]repositories.MemorySearchResultRow, error)
}

type Service struct {
	membershipChecker MembershipChecker
	repositoryReader  RepositoryReader
	memorySearcher    MemorySearcher
}

func NewService(membershipChecker MembershipChecker, repositoryReader RepositoryReader, memorySearcher MemorySearcher) *Service {
	return &Service{
		membershipChecker: membershipChecker,
		repositoryReader:  repositoryReader,
		memorySearcher:    memorySearcher,
	}
}

type MemorySearchInput struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	RepositoryID   *uuid.UUID
	Query          string
	Page           int
	PageSize       int
}

type MemorySearchResult struct {
	ID             uuid.UUID
	RepositoryID   uuid.UUID
	RepositoryName string
	Type           string
	Title          string
	SummarySnippet string
	SourceURL      string
	CreatedAt      time.Time
}

type MemorySearchResponse struct {
	Query    string
	Page     int
	PageSize int
	Total    int
	Results  []MemorySearchResult
}

func (s *Service) SearchMemory(ctx context.Context, input MemorySearchInput) (MemorySearchResponse, error) {
	if err := s.authorize(ctx, input.UserID, input.OrganizationID, input.RepositoryID); err != nil {
		return MemorySearchResponse{}, err
	}

	page := input.Page
	if page < 1 {
		page = DefaultPage
	}
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	query := strings.TrimSpace(input.Query)
	if query == "" {
		return MemorySearchResponse{
			Query:    "",
			Page:     page,
			PageSize: pageSize,
			Total:    0,
			Results:  []MemorySearchResult{},
		}, nil
	}

	offset := (page - 1) * pageSize
	rows, err := s.memorySearcher.Search(ctx, repositories.MemorySearchInput{
		OrganizationID: input.OrganizationID,
		RepositoryID:   input.RepositoryID,
		Query:          query,
		Limit:          int32(pageSize),
		Offset:         int32(offset),
	})
	if err != nil {
		return MemorySearchResponse{}, err
	}

	total := 0
	if len(rows) > 0 {
		total = int(rows[0].TotalCount)
	}

	results := make([]MemorySearchResult, 0, len(rows))
	for _, row := range rows {
		results = append(results, MemorySearchResult{
			ID:             row.ID,
			RepositoryID:   row.RepositoryID,
			RepositoryName: row.RepositoryName,
			Type:           row.Type,
			Title:          row.Title,
			SummarySnippet: toSnippet(row.Summary, 180),
			SourceURL:      row.SourceURL,
			CreatedAt:      row.CreatedAt,
		})
	}

	return MemorySearchResponse{
		Query:    query,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Results:  results,
	}, nil
}

func (s *Service) authorize(ctx context.Context, userID, organizationID uuid.UUID, repositoryID *uuid.UUID) error {
	allowed, err := s.membershipChecker.UserHasMembership(ctx, db.UserHasMembershipParams{
		UserID:         userID,
		OrganizationID: organizationID,
	})
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}

	if repositoryID == nil {
		return nil
	}

	repo, err := s.repositoryReader.GetByID(ctx, *repositoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if repo.OrganizationID != organizationID {
		return ErrForbidden
	}

	return nil
}

func toSnippet(summary string, max int) string {
	summary = strings.Join(strings.Fields(strings.TrimSpace(summary)), " ")
	if len(summary) <= max {
		return summary
	}
	if max <= 3 {
		return summary[:max]
	}
	return strings.TrimSpace(summary[:max-3]) + "..."
}
