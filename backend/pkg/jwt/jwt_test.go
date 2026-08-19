package jwt

import (
	"testing"
	"time"
)

func TestGenerateParseRoundTrip(t *testing.T) {
	token, jti, err := Generate("unit-secret", time.Hour, 42, "alice")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if jti == "" {
		t.Fatal("jti empty")
	}
	claims, err := Parse("unit-secret", token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != 42 || claims.Username != "alice" {
		t.Fatalf("claims mismatch: %+v", claims)
	}
	if claims.ID != jti {
		t.Fatalf("jti mismatch: got %s want %s", claims.ID, jti)
	}
	if claims.Issuer != "dxcloud" {
		t.Fatalf("issuer mismatch: %s", claims.Issuer)
	}
}

func TestParseWrongSecret(t *testing.T) {
	token, _, _ := Generate("secret-a", time.Hour, 1, "u")
	if _, err := Parse("secret-b", token); err == nil {
		t.Fatal("expected error with wrong secret")
	}
}

func TestParseExpired(t *testing.T) {
	token, _, _ := Generate("secret-a", -time.Minute, 1, "u")
	if _, err := Parse("secret-a", token); err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestParseGarbage(t *testing.T) {
	if _, err := Parse("secret-a", "not-a-jwt"); err == nil {
		t.Fatal("expected error for garbage token")
	}
}
