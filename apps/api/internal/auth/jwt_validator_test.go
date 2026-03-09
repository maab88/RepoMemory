package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTIdentityValidatorValidateBearerToken(t *testing.T) {
	validator, err := NewJWTIdentityValidator(JWTValidatorConfig{
		Secret:   "test-secret",
		Issuer:   "repomemory-web",
		Audience: "repomemory-api",
	})
	if err != nil {
		t.Fatalf("new validator: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":     "repomemory-web",
		"sub":     "github|123",
		"aud":     "repomemory-api",
		"email":   "user@example.com",
		"name":    "Jane Doe",
		"picture": "https://example.com/avatar.png",
		"exp":     time.Now().Add(5 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	})
	signed, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	identity, err := validator.ValidateBearerToken(context.Background(), signed)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}

	if identity.Subject != "github|123" {
		t.Fatalf("expected subject github|123, got %s", identity.Subject)
	}
}

func TestJWTIdentityValidatorRejectsWrongSecret(t *testing.T) {
	validator, err := NewJWTIdentityValidator(JWTValidatorConfig{
		Secret:   "secret-a",
		Issuer:   "repomemory-web",
		Audience: "repomemory-api",
	})
	if err != nil {
		t.Fatalf("new validator: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "repomemory-web",
		"sub": "github|123",
		"aud": "repomemory-api",
		"exp": time.Now().Add(5 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	})
	signed, err := token.SignedString([]byte("secret-b"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := validator.ValidateBearerToken(context.Background(), signed); err == nil {
		t.Fatal("expected invalid auth token error")
	}
}
