package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/auth"
	"github.com/maab88/repomemory/apps/api/internal/http/dto"
	"github.com/maab88/repomemory/apps/api/internal/http/response"
	servicejobs "github.com/maab88/repomemory/apps/api/internal/services/jobs"
)

func (h *V1Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.CurrentUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing current user")
		return
	}
	if h.jobService == nil {
		response.WriteError(w, http.StatusServiceUnavailable, "service_unavailable", "job service is not configured")
		return
	}

	jobID, err := uuid.Parse(chi.URLParam(r, "jobId"))
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "bad_request", "invalid job id")
		return
	}

	job, err := h.jobService.GetJob(r.Context(), currentUser.ID, jobID)
	if err != nil {
		switch {
		case errors.Is(err, servicejobs.ErrJobForbidden):
			response.WriteError(w, http.StatusForbidden, "forbidden", "access denied")
		case errors.Is(err, servicejobs.ErrJobNotFound):
			response.WriteError(w, http.StatusNotFound, "not_found", "job not found")
		default:
			response.WriteError(w, http.StatusInternalServerError, "internal_error", "failed to fetch job")
		}
		return
	}

	response.WriteData(w, http.StatusOK, map[string]any{
		"job": dto.ToJobDTO(job),
	})
}
