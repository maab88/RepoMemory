package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/maab88/repomemory/apps/api/internal/db"
)

type fakeUserUpserter struct {
	lastArg db.UpsertUserByIDParams
	user    db.User
}

func (f *fakeUserUpserter) UpsertUserByID(_ context.Context, arg db.UpsertUserByIDParams) (db.User, error) {
	f.lastArg = arg
	if f.user.ID == uuid.Nil {
		f.user = db.User{
			ID:          arg.ID,
			Email:       arg.Email,
			DisplayName: arg.DisplayName,
			AvatarUrl:   arg.AvatarUrl,
			CreatedAt:   pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		}
	}
	return f.user, nil
}

func TestUserMapperCreatesStableLocalUser(t *testing.T) {
	upserter := &fakeUserUpserter{}
	mapper := NewUserMapper(upserter)
	identity := ExternalIdentity{
		Subject:     "auth0|abc123",
		Issuer:      "repomemory-web",
		Email:       "user@example.com",
		DisplayName: "Jane Doe",
		AvatarURL:   "https://example.com/avatar.png",
	}

	user, err := mapper.MapToCurrentUser(context.Background(), identity)
	if err != nil {
		t.Fatalf("map user: %v", err)
	}

	expectedID := uuid.NewSHA1(uuid.NameSpaceOID, []byte(identity.Issuer+"|"+identity.Subject))
	if user.ID != expectedID {
		t.Fatalf("expected stable id %s, got %s", expectedID, user.ID)
	}
	if upserter.lastArg.DisplayName != "Jane Doe" {
		t.Fatalf("expected display name to be passed through")
	}
	if !upserter.lastArg.Email.Valid || upserter.lastArg.Email.String != "user@example.com" {
		t.Fatalf("expected email to be persisted")
	}
}

func TestUserMapperMapsRepeatLoginToSameID(t *testing.T) {
	upserter := &fakeUserUpserter{}
	mapper := NewUserMapper(upserter)
	identity := ExternalIdentity{Subject: "sub-42", Issuer: "repomemory-web"}

	first, err := mapper.MapToCurrentUser(context.Background(), identity)
	if err != nil {
		t.Fatalf("first map: %v", err)
	}
	second, err := mapper.MapToCurrentUser(context.Background(), identity)
	if err != nil {
		t.Fatalf("second map: %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("expected same local user id on repeat login")
	}
}
