package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/maab88/repomemory/apps/api/internal/db"
	"github.com/maab88/repomemory/apps/api/internal/repositories"
)

var (
	ErrRepositoryForbidden = errors.New("repository access denied")
	ErrRepositoryNotFound  = errors.New("repository not found")
	ErrMemoryEntryNotFound = errors.New("memory entry not found")
)

type MembershipChecker interface {
	UserHasMembership(ctx context.Context, arg db.UserHasMembershipParams) (bool, error)
}

type QueryService struct {
	repositoryRepository        *repositories.RepositoryRepository
	memoryEntryRepository       *repositories.MemoryEntryRepository
	memoryEntrySourceRepository *repositories.MemoryEntrySourceRepository
	membershipChecker           MembershipChecker
}

func NewQueryService(
	repositoryRepository *repositories.RepositoryRepository,
	memoryEntryRepository *repositories.MemoryEntryRepository,
	memoryEntrySourceRepository *repositories.MemoryEntrySourceRepository,
	membershipChecker MembershipChecker,
) *QueryService {
	return &QueryService{
		repositoryRepository:        repositoryRepository,
		memoryEntryRepository:       memoryEntryRepository,
		memoryEntrySourceRepository: memoryEntrySourceRepository,
		membershipChecker:           membershipChecker,
	}
}

type MemoryEntry struct {
	ID             uuid.UUID
	RepositoryID   uuid.UUID
	OrganizationID uuid.UUID
	Type           string
	Title          string
	Summary        string
	WhyItMatters   string
	ImpactedAreas  []string
	Risks          []string
	FollowUps      []string
	GeneratedBy    string
	SourceURL      string
	CreatedAt      time.Time
	Sources        []MemorySource
}

type MemorySource struct {
	SourceType   string
	SourceURL    string
	DisplayLabel string
}

func (s *QueryService) ListRepositoryMemory(ctx context.Context, userID, repositoryID uuid.UUID) ([]MemoryEntry, error) {
	repo, err := s.authorizeRepository(ctx, userID, repositoryID)
	if err != nil {
		return nil, err
	}

	rows, err := s.memoryEntryRepository.ListByRepository(ctx, repositoryID)
	if err != nil {
		return nil, err
	}

	result := make([]MemoryEntry, 0, len(rows))
	for _, row := range rows {
		entry, err := mapMemoryEntry(row)
		if err != nil {
			return nil, err
		}
		entry.OrganizationID = repo.OrganizationID
		result = append(result, entry)
	}
	return result, nil
}

func (s *QueryService) GetRepositoryMemoryEntry(ctx context.Context, userID, repositoryID, memoryID uuid.UUID) (MemoryEntry, error) {
	repo, err := s.authorizeRepository(ctx, userID, repositoryID)
	if err != nil {
		return MemoryEntry{}, err
	}

	row, err := s.memoryEntryRepository.GetByIDAndRepository(ctx, memoryID, repositoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return MemoryEntry{}, ErrMemoryEntryNotFound
		}
		return MemoryEntry{}, err
	}

	entry, err := mapMemoryEntry(row)
	if err != nil {
		return MemoryEntry{}, err
	}
	entry.OrganizationID = repo.OrganizationID

	sourceRows, err := s.memoryEntrySourceRepository.ListByMemoryEntryID(ctx, memoryID)
	if err != nil {
		return MemoryEntry{}, err
	}
	sources := make([]MemorySource, 0, len(sourceRows))
	for _, source := range sourceRows {
		sources = append(sources, MemorySource{
			SourceType:   source.SourceType,
			SourceURL:    source.SourceUrl,
			DisplayLabel: sourceDisplayLabel(source.SourceType, source.SourceUrl),
		})
	}
	entry.Sources = sources
	return entry, nil
}

func (s *QueryService) authorizeRepository(ctx context.Context, userID, repositoryID uuid.UUID) (db.Repository, error) {
	repo, err := s.repositoryRepository.GetByID(ctx, repositoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Repository{}, ErrRepositoryNotFound
		}
		return db.Repository{}, err
	}
	allowed, err := s.membershipChecker.UserHasMembership(ctx, db.UserHasMembershipParams{
		UserID:         userID,
		OrganizationID: repo.OrganizationID,
	})
	if err != nil {
		return db.Repository{}, err
	}
	if !allowed {
		return db.Repository{}, ErrRepositoryForbidden
	}
	return repo, nil
}

func mapMemoryEntry(row db.MemoryEntry) (MemoryEntry, error) {
	impacted, err := decodeStringArray(row.ImpactedAreas)
	if err != nil {
		return MemoryEntry{}, err
	}
	risks, err := decodeStringArray(row.Risks)
	if err != nil {
		return MemoryEntry{}, err
	}
	followUps, err := decodeStringArray(row.FollowUps)
	if err != nil {
		return MemoryEntry{}, err
	}

	out := MemoryEntry{
		ID:             row.ID,
		RepositoryID:   row.RepositoryID,
		OrganizationID: row.OrganizationID,
		Type:           row.Type,
		Title:          row.Title,
		Summary:        row.Summary,
		WhyItMatters:   row.WhyItMatters.String,
		ImpactedAreas:  impacted,
		Risks:          risks,
		FollowUps:      followUps,
		GeneratedBy:    row.GeneratedBy,
		SourceURL:      row.SourceUrl.String,
		CreatedAt:      row.CreatedAt.Time.UTC(),
	}
	return out, nil
}

func decodeStringArray(raw []byte) ([]string, error) {
	if len(raw) == 0 {
		return []string{}, nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if out == nil {
		return []string{}, nil
	}
	return out, nil
}

func sourceDisplayLabel(sourceType, sourceURL string) string {
	prefix := "Source"
	switch sourceType {
	case "pull_request":
		prefix = "PR"
	case "issue":
		prefix = "Issue"
	}

	if sourceURL != "" {
		if n := trailingNumber(sourceURL); n != "" {
			return fmt.Sprintf("%s #%s", prefix, n)
		}
	}
	return prefix
}

func trailingNumber(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	for _, ch := range last {
		if ch < '0' || ch > '9' {
			return ""
		}
	}
	return last
}
