package auth

import (
	"net/http"

	"github.com/maab88/repomemory/apps/api/internal/http/response"
	"github.com/rs/zerolog/log"
)

func RequireMockAuth(resolver UserResolver) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := headerOrEmpty(r.Header, "x-user-id")
			if userID == "" {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing x-user-id header")
				return
			}

			user, err := resolver.Resolve(r.Context(), MockUserInput{
				RawID: userID,
				Email: headerOrEmpty(r.Header, "x-user-email"),
				Name:  headerOrEmpty(r.Header, "x-user-name"),
			})
			if err != nil {
				log.Error().Err(err).Str("user_id", userID).Msg("mock auth resolver failed")
				response.WriteError(w, http.StatusUnauthorized, "unauthorized", "failed to resolve current user")
				return
			}

			next.ServeHTTP(w, r.WithContext(WithCurrentUser(r.Context(), user)))
		})
	}
}
