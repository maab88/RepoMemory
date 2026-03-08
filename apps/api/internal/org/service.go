package org

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrInvalidOrganizationName = errors.New("invalid organization name")
	ErrOrganizationConflict    = errors.New("organization already exists")
	ErrOrganizationNotFound    = errors.New("organization not found")
	ErrOrganizationForbidden   = errors.New("organization access denied")
)

type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) CreateOrganization(ctx context.Context, userID uuid.UUID, name string) (OrganizationWithRole, error) {
	cleanName := strings.TrimSpace(name)
	if len(cleanName) < 2 || len(cleanName) > 80 {
		return OrganizationWithRole{}, ErrInvalidOrganizationName
	}

	slug := Slugify(cleanName)
	if len(slug) < 2 {
		return OrganizationWithRole{}, ErrInvalidOrganizationName
	}

	org, err := s.store.CreateOrganizationWithOwner(ctx, userID, cleanName, slug)
	if err != nil {
		if IsUniqueViolation(err) {
			return OrganizationWithRole{}, ErrOrganizationConflict
		}
		return OrganizationWithRole{}, err
	}
	return org, nil
}

func (s *Service) ListOrganizations(ctx context.Context, userID uuid.UUID) ([]OrganizationWithRole, error) {
	return s.store.ListOrganizationsWithRoleForUser(ctx, userID)
}

func (s *Service) GetOrganization(ctx context.Context, userID, orgID uuid.UUID) (OrganizationWithRole, error) {
	org, err := s.store.GetOrganizationForUser(ctx, orgID, userID)
	if err == nil {
		return org, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return OrganizationWithRole{}, err
	}

	_, existsErr := s.store.GetOrganizationByID(ctx, orgID)
	if existsErr == nil {
		return OrganizationWithRole{}, ErrOrganizationForbidden
	}
	if errors.Is(existsErr, pgx.ErrNoRows) {
		return OrganizationWithRole{}, ErrOrganizationNotFound
	}
	return OrganizationWithRole{}, existsErr
}
