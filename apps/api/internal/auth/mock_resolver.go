package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type MockUserInput struct {
	RawID string
	Email string
	Name  string
}

type UserResolver interface {
	Resolve(ctx context.Context, input MockUserInput) (CurrentUser, error)
}

type MockUserResolver struct {
	queries *db.Queries
}

func NewMockUserResolver(queries *db.Queries) *MockUserResolver {
	return &MockUserResolver{queries: queries}
}

func (r *MockUserResolver) Resolve(ctx context.Context, input MockUserInput) (CurrentUser, error) {
	id := resolveID(input.RawID)
	name := input.Name
	if name == "" {
		name = fmt.Sprintf("Dev User %s", id.String()[:8])
	}

	email := pgtype.Text{}
	if input.Email != "" {
		email = pgtype.Text{String: input.Email, Valid: true}
	}

	user, err := r.queries.UpsertUserByID(ctx, db.UpsertUserByIDParams{
		ID:          id,
		Email:       email,
		DisplayName: name,
		AvatarUrl:   pgtype.Text{},
	})
	if err != nil {
		return CurrentUser{}, err
	}

	out := CurrentUser{ID: user.ID, DisplayName: user.DisplayName}
	if user.Email.Valid {
		out.Email = user.Email.String
	}
	if user.AvatarUrl.Valid {
		out.AvatarURL = user.AvatarUrl.String
	}
	return out, nil
}

func resolveID(raw string) uuid.UUID {
	parsed, err := uuid.Parse(raw)
	if err == nil {
		return parsed
	}
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(raw))
}

func headerOrEmpty(h http.Header, key string) string {
	return h.Get(key)
}
