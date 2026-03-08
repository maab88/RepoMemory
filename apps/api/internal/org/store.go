package org

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type OrganizationWithRole struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
	Role string    `json:"role"`
}

type Store interface {
	CreateOrganizationWithOwner(ctx context.Context, userID uuid.UUID, name, slug string) (OrganizationWithRole, error)
	ListOrganizationsWithRoleForUser(ctx context.Context, userID uuid.UUID) ([]OrganizationWithRole, error)
	GetOrganizationForUser(ctx context.Context, orgID, userID uuid.UUID) (OrganizationWithRole, error)
	GetOrganizationByID(ctx context.Context, orgID uuid.UUID) (db.Organization, error)
}

type PGStore struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

func NewStore(pool *pgxpool.Pool, queries *db.Queries) *PGStore {
	return &PGStore{pool: pool, queries: queries}
}

func (s *PGStore) CreateOrganizationWithOwner(ctx context.Context, userID uuid.UUID, name, slug string) (OrganizationWithRole, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return OrganizationWithRole{}, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := s.queries.WithTx(tx)
	orgRow, err := qtx.CreateOrganization(ctx, db.CreateOrganizationParams{Name: name, Slug: slug})
	if err != nil {
		return OrganizationWithRole{}, err
	}

	if _, err := qtx.CreateMembership(ctx, db.CreateMembershipParams{
		OrganizationID: orgRow.ID,
		UserID:         userID,
		Role:           "owner",
	}); err != nil {
		return OrganizationWithRole{}, err
	}

	metadata, err := json.Marshal(map[string]any{"name": name, "slug": slug})
	if err != nil {
		return OrganizationWithRole{}, err
	}

	if _, err := qtx.InsertAuditLog(ctx, db.InsertAuditLogParams{
		OrganizationID: &orgRow.ID,
		ActorUserID:    &userID,
		Action:         "organization.created",
		EntityType:     "organization",
		EntityID:       &orgRow.ID,
		Metadata:       metadata,
	}); err != nil {
		return OrganizationWithRole{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return OrganizationWithRole{}, err
	}

	return OrganizationWithRole{ID: orgRow.ID, Name: orgRow.Name, Slug: orgRow.Slug, Role: "owner"}, nil
}

func (s *PGStore) ListOrganizationsWithRoleForUser(ctx context.Context, userID uuid.UUID) ([]OrganizationWithRole, error) {
	rows, err := s.queries.ListOrganizationsWithRoleForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]OrganizationWithRole, 0, len(rows))
	for _, row := range rows {
		out = append(out, OrganizationWithRole{ID: row.ID, Name: row.Name, Slug: row.Slug, Role: row.Role})
	}
	return out, nil
}

func (s *PGStore) GetOrganizationForUser(ctx context.Context, orgID, userID uuid.UUID) (OrganizationWithRole, error) {
	row, err := s.queries.GetOrganizationForUser(ctx, db.GetOrganizationForUserParams{ID: orgID, UserID: userID})
	if err != nil {
		return OrganizationWithRole{}, err
	}
	return OrganizationWithRole{ID: row.ID, Name: row.Name, Slug: row.Slug, Role: row.Role}, nil
}

func (s *PGStore) GetOrganizationByID(ctx context.Context, orgID uuid.UUID) (db.Organization, error) {
	return s.queries.GetOrganizationByID(ctx, orgID)
}

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func Slugify(name string) string {
	trimmed := strings.TrimSpace(strings.ToLower(name))
	parts := strings.Fields(trimmed)
	slug := strings.Join(parts, "-")
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)
	return strings.Trim(slug, "-")
}
