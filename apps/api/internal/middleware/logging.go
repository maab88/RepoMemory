package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/maab88/repomemory/apps/api/internal/auth"
	"github.com/rs/zerolog/log"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		event := log.Info().
			Str("request_id", RequestIDFromContext(r.Context())).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", recorder.status).
			Int64("duration_ms", time.Since(start).Milliseconds())

		if currentUser, ok := auth.CurrentUserFromContext(r.Context()); ok {
			event = event.Str("user_id", currentUser.ID.String())
		}

		if orgID := parseOrganizationID(chi.URLParam(r, "orgId")); orgID != "" {
			event = event.Str("organization_id", orgID)
		}

		event.Msg("api request completed")
	})
}

func parseOrganizationID(raw string) string {
	if raw == "" {
		return ""
	}
	if _, err := uuid.Parse(raw); err != nil {
		return ""
	}
	return raw
}
