package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/auth"
	"github.com/maab88/repomemory/apps/api/internal/http/dto"
	"github.com/maab88/repomemory/apps/api/internal/http/response"
	servicerepositories "github.com/maab88/repomemory/apps/api/internal/services/repositories"
)

func (h *V1Handler) ListRepositoryDigests(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}
	if h.repositoryService == nil {
		response.WriteError(w, http.StatusServiceUnavailable, "service_unavailable", "repository service is not configured")
		return
	}

	repositoryID, err := uuid.Parse(chi.URLParam(r, "repoId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid repository id")
		return
	}

	digests, err := h.repositoryService.ListDigests(r.Context(), currentUser.ID, repositoryID)
	if err != nil {
		switch {
		case errors.Is(err, servicerepositories.ErrRepositoryForbidden):
			response.WriteError(w, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, servicerepositories.ErrRepositoryNotFound):
			response.WriteError(w, http.StatusNotFound, "not_found", "repository not found")
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list digests")
		}
		return
	}

	items := make([]dto.DigestDTO, 0, len(digests))
	for _, item := range digests {
		items = append(items, dto.ToDigestDTO(item))
	}

	response.WriteData(w, http.StatusOK, map[string]any{
		"digests": items,
	})
}

func (h *V1Handler) GenerateRepositoryDigest(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}
	if h.repositoryService == nil {
		response.WriteError(w, http.StatusServiceUnavailable, "service_unavailable", "repository service is not configured")
		return
	}

	repositoryID, err := uuid.Parse(chi.URLParam(r, "repoId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid repository id")
		return
	}

	job, err := h.repositoryService.TriggerDigestGeneration(r.Context(), currentUser.ID, repositoryID)
	if err != nil {
		switch {
		case errors.Is(err, servicerepositories.ErrRepositoryForbidden):
			response.WriteError(w, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, servicerepositories.ErrRepositoryNotFound):
			response.WriteError(w, http.StatusNotFound, "not_found", "repository not found")
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to trigger digest generation")
		}
		return
	}

	response.WriteData(w, http.StatusAccepted, map[string]any{
		"jobId":  job.ID,
		"status": job.Status,
	})
}
