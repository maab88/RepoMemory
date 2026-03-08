package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/auth"
	"github.com/maab88/repomemory/apps/api/internal/http/dto"
	"github.com/maab88/repomemory/apps/api/internal/http/response"
	servicememory "github.com/maab88/repomemory/apps/api/internal/services/memory"
)

func (h *V1Handler) ListRepositoryMemory(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}
	if h.memoryService == nil {
		response.WriteError(w, http.StatusServiceUnavailable, "service_unavailable", "memory service is not configured")
		return
	}

	repositoryID, err := uuid.Parse(chi.URLParam(r, "repoId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid repository id")
		return
	}

	entries, err := h.memoryService.ListRepositoryMemory(r.Context(), currentUser.ID, repositoryID)
	if err != nil {
		switch {
		case errors.Is(err, servicememory.ErrRepositoryForbidden):
			response.WriteError(w, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, servicememory.ErrRepositoryNotFound):
			response.WriteError(w, http.StatusNotFound, "not_found", "repository not found")
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to list repository memory")
		}
		return
	}

	items := make([]dto.MemoryEntryDTO, 0, len(entries))
	for _, entry := range entries {
		items = append(items, dto.ToMemoryEntryDTO(entry))
	}

	response.WriteData(w, http.StatusOK, map[string]any{
		"memoryEntries": items,
	})
}

func (h *V1Handler) GetRepositoryMemoryDetail(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}
	if h.memoryService == nil {
		response.WriteError(w, http.StatusServiceUnavailable, "service_unavailable", "memory service is not configured")
		return
	}

	repositoryID, err := uuid.Parse(chi.URLParam(r, "repoId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid repository id")
		return
	}
	memoryID, err := uuid.Parse(chi.URLParam(r, "memoryId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid memory id")
		return
	}

	entry, err := h.memoryService.GetRepositoryMemoryEntry(r.Context(), currentUser.ID, repositoryID, memoryID)
	if err != nil {
		switch {
		case errors.Is(err, servicememory.ErrRepositoryForbidden):
			response.WriteError(w, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, servicememory.ErrRepositoryNotFound), errors.Is(err, servicememory.ErrMemoryEntryNotFound):
			response.WriteError(w, http.StatusNotFound, "not_found", "memory entry not found")
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to load memory entry")
		}
		return
	}

	response.WriteData(w, http.StatusOK, map[string]any{
		"memoryEntry": dto.ToMemoryEntryDetailDTO(entry),
	})
}
