package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/auth"
	"github.com/maab88/repomemory/apps/api/internal/http/dto"
	"github.com/maab88/repomemory/apps/api/internal/http/response"
	servicesearch "github.com/maab88/repomemory/apps/api/internal/services/search"
)

func (h *V1Handler) SearchMemory(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}
	if h.searchService == nil {
		response.WriteError(w, http.StatusServiceUnavailable, "service_unavailable", "search service is not configured")
		return
	}

	organizationID, err := uuid.Parse(strings.TrimSpace(r.URL.Query().Get("organizationId")))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid organization id")
		return
	}

	var repositoryID *uuid.UUID
	repositoryIDRaw := strings.TrimSpace(r.URL.Query().Get("repositoryId"))
	if repositoryIDRaw != "" {
		parsed, parseErr := uuid.Parse(repositoryIDRaw)
		if parseErr != nil {
			response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid repository id")
			return
		}
		repositoryID = &parsed
	}

	page := 1
	if raw := strings.TrimSpace(r.URL.Query().Get("page")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil {
			response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid page")
			return
		}
		page = parsed
	}

	pageSize := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("pageSize")); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil {
			response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid page size")
			return
		}
		pageSize = parsed
	}

	result, err := h.searchService.SearchMemory(r.Context(), servicesearch.MemorySearchInput{
		UserID:         currentUser.ID,
		OrganizationID: organizationID,
		RepositoryID:   repositoryID,
		Query:          r.URL.Query().Get("q"),
		Page:           page,
		PageSize:       pageSize,
	})
	if err != nil {
		switch {
		case errors.Is(err, servicesearch.ErrForbidden):
			response.WriteError(w, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, servicesearch.ErrNotFound):
			response.WriteError(w, http.StatusNotFound, "not_found", "repository not found")
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to search memory")
		}
		return
	}

	response.WriteData(w, http.StatusOK, dto.ToMemorySearchDataDTO(result))
}
