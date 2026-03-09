package auth

import (
	"net/http"
	"strings"

	"github.com/maab88/repomemory/apps/api/internal/http/response"
	"github.com/rs/zerolog/log"
)

func RequireBearerAuth(validator IdentityValidator, mapper IdentityMapper) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := parseBearerToken(r.Header.Get("Authorization"))
			if !ok {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid authorization bearer token")
				return
			}

			identity, err := validator.ValidateBearerToken(r.Context(), token)
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid or expired authorization token")
				return
			}

			user, err := mapper.MapToCurrentUser(r.Context(), identity)
			if err != nil {
				log.Error().Err(err).Str("subject", identity.Subject).Str("issuer", identity.Issuer).Msg("failed to map external identity")
				response.WriteError(w, http.StatusUnauthorized, "unauthorized", "failed to resolve current user")
				return
			}

			next.ServeHTTP(w, r.WithContext(WithCurrentUser(r.Context(), user)))
		})
	}
}

func parseBearerToken(raw string) (string, bool) {
	parts := strings.SplitN(strings.TrimSpace(raw), " ", 2)
	if len(parts) != 2 {
		return "", false
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	return token, token != ""
}
