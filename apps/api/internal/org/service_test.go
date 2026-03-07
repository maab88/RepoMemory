package org

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type fakeStore struct {
	createFn func(ctx context.Context, userID uuid.UUID, name, slug string) (OrganizationWithRole, error)
	listFn   func(ctx context.Context, userID uuid.UUID) ([]OrganizationWithRole, error)
	getFn    func(ctx context.Context, orgID, userID uuid.UUID) (OrganizationWithRole, error)
	getByID  func(ctx context.Context, orgID uuid.UUID) (db.Organization, error)
}

func (f *fakeStore) CreateOrganizationWithOwner(ctx context.Context, userID uuid.UUID, name, slug string) (OrganizationWithRole, error) {
	return f.createFn(ctx, userID, name, slug)
}

func (f *fakeStore) ListOrganizationsWithRoleForUser(ctx context.Context, userID uuid.UUID) ([]OrganizationWithRole, error) {
	return f.listFn(ctx, userID)
}

func (f *fakeStore) GetOrganizationForUser(ctx context.Context, orgID, userID uuid.UUID) (OrganizationWithRole, error) {
	return f.getFn(ctx, orgID, userID)
}

func (f *fakeStore) GetOrganizationByID(ctx context.Context, orgID uuid.UUID) (db.Organization, error) {
	return f.getByID(ctx, orgID)
}

func TestCreateOrganizationSuccess(t *testing.T) {
	store := &fakeStore{
		createFn: func(_ context.Context, _ uuid.UUID, name, slug string) (OrganizationWithRole, error) {
			if name != "Acme Inc" {
				t.Fatalf("unexpected name: %s", name)
			}
			if slug != "acme-inc" {
				t.Fatalf("unexpected slug: %s", slug)
			}
			return OrganizationWithRole{ID: uuid.New(), Name: name, Slug: slug, Role: "owner"}, nil
		},
	}

	svc := NewService(store)
	out, err := svc.CreateOrganization(context.Background(), uuid.New(), "Acme Inc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Role != "owner" {
		t.Fatalf("expected owner role, got %s", out.Role)
	}
}

func TestCreateOrganizationInvalidName(t *testing.T) {
	svc := NewService(&fakeStore{})
	_, err := svc.CreateOrganization(context.Background(), uuid.New(), " ")
	if !errors.Is(err, ErrInvalidOrganizationName) {
		t.Fatalf("expected invalid name error, got %v", err)
	}
}

func TestCreateOrganizationDuplicateSlug(t *testing.T) {
	store := &fakeStore{
		createFn: func(_ context.Context, _ uuid.UUID, _ string, _ string) (OrganizationWithRole, error) {
			return OrganizationWithRole{}, &pgconn.PgError{Code: "23505"}
		},
	}
	svc := NewService(store)

	_, err := svc.CreateOrganization(context.Background(), uuid.New(), "Acme")
	if !errors.Is(err, ErrOrganizationConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestGetOrganizationAccessDenied(t *testing.T) {
	orgID := uuid.New()
	store := &fakeStore{
		getFn: func(_ context.Context, _, _ uuid.UUID) (OrganizationWithRole, error) {
			return OrganizationWithRole{}, pgx.ErrNoRows
		},
		getByID: func(_ context.Context, _ uuid.UUID) (db.Organization, error) {
			return db.Organization{ID: orgID, Name: "Acme", Slug: "acme"}, nil
		},
	}
	service := NewService(store)

	_, err := service.GetOrganization(context.Background(), uuid.New(), orgID)
	if !errors.Is(err, ErrOrganizationForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}
