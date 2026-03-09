package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type UserUpserter interface {
	UpsertUserByID(ctx context.Context, arg db.UpsertUserByIDParams) (db.User, error)
}

type UserMapper struct {
	users UserUpserter
}

func NewUserMapper(users UserUpserter) *UserMapper {
	return &UserMapper{users: users}
}

func (m *UserMapper) MapToCurrentUser(ctx context.Context, identity ExternalIdentity) (CurrentUser, error) {
	stableID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(identity.Issuer+"|"+identity.Subject))
	displayName := identity.DisplayName
	if strings.TrimSpace(displayName) == "" {
		displayName = fmt.Sprintf("User %s", stableID.String()[:8])
	}

	user, err := m.users.UpsertUserByID(ctx, db.UpsertUserByIDParams{
		ID:          stableID,
		Email:       toText(identity.Email),
		DisplayName: displayName,
		AvatarUrl:   toText(identity.AvatarURL),
	})
	if err != nil {
		return CurrentUser{}, err
	}

	current := CurrentUser{
		ID:          user.ID,
		DisplayName: user.DisplayName,
	}
	if user.Email.Valid {
		current.Email = user.Email.String
	}
	if user.AvatarUrl.Valid {
		current.AvatarURL = user.AvatarUrl.String
	}
	if user.CreatedAt.Valid {
		current.CreatedAt = user.CreatedAt.Time
	}

	return current, nil
}

func toText(value string) pgtype.Text {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: trimmed, Valid: true}
}
