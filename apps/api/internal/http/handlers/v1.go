package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/auth"
	"github.com/maab88/repomemory/apps/api/internal/http/response"
	"github.com/maab88/repomemory/apps/api/internal/org"
)

type OrganizationService interface {
	CreateOrganization(ctx context.Context, userID uuid.UUID, name string) (org.OrganizationWithRole, error)
	ListOrganizations(ctx context.Context, userID uuid.UUID) ([]org.OrganizationWithRole, error)
	GetOrganization(ctx context.Context, userID, orgID uuid.UUID) (org.OrganizationWithRole, error)
}

type V1Handler struct {
	orgService OrganizationService
}

func NewV1Handler(orgService OrganizationService) *V1Handler {
	return &V1Handler{orgService: orgService}
}

type meResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email,omitempty"`
	DisplayName string    `json:"displayName"`
	AvatarURL   string    `json:"avatarUrl,omitempty"`
}

func (h *V1Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}

	response.WriteData(w, http.StatusOK, meResponse{
		ID:          currentUser.ID,
		Email:       currentUser.Email,
		DisplayName: currentUser.DisplayName,
		AvatarURL:   currentUser.AvatarURL,
	})
}

func (h *V1Handler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}

	orgs, err := h.orgService.ListOrganizations(r.Context(), currentUser.ID)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list organizations")
		return
	}

	response.WriteData(w, http.StatusOK, orgs)
}

type createOrganizationRequest struct {
	Name string `json:"name"`
}

func (h *V1Handler) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}

	var req createOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid request payload")
		return
	}

	created, err := h.orgService.CreateOrganization(r.Context(), currentUser.ID, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, org.ErrInvalidOrganizationName):
			response.WriteError(w, http.StatusBadRequest, "validation_error", "organization name must be 2-80 characters")
		case errors.Is(err, org.ErrOrganizationConflict):
			response.WriteError(w, http.StatusConflict, "conflict", "organization slug already exists")
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to create organization")
		}
		return
	}

	response.WriteData(w, http.StatusCreated, created)
}

func (h *V1Handler) GetOrganization(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}

	orgIDRaw := chi.URLParam(r, "orgID")
	orgID, err := uuid.Parse(orgIDRaw)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid organization id")
		return
	}

	orgData, err := h.orgService.GetOrganization(r.Context(), currentUser.ID, orgID)
	if err != nil {
		switch {
		case errors.Is(err, org.ErrOrganizationForbidden):
			response.WriteError(w, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, org.ErrOrganizationNotFound):
			response.WriteError(w, http.StatusNotFound, "not_found", "organization not found")
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to fetch organization")
		}
		return
	}

	response.WriteData(w, http.StatusOK, orgData)
}
