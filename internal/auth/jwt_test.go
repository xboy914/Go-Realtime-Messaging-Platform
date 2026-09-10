package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-with-at-least-thirty-two-bytes"

func TestIssueAndVerify(t *testing.T) {
	manager, err := NewManager(testSecret, "messaging-test", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	token, err := manager.Issue("user-42", "Demo User")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "user-42" || claims.DisplayName != "Demo User" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestRejectsExpiredToken(t *testing.T) {
	manager, err := NewManager(testSecret, "messaging-test", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	claims := Claims{RegisteredClaims: jwt.RegisteredClaims{
		Subject: "user-42", Issuer: "messaging-test",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
	}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(manager.secret)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Verify(token); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}

func TestExtractBearer(t *testing.T) {
	token, err := ExtractBearer("Bearer signed-token")
	if err != nil || token != "signed-token" {
		t.Fatalf("unexpected result: token=%q err=%v", token, err)
	}
	if _, err := ExtractBearer("signed-token"); err == nil {
		t.Fatal("expected missing Bearer scheme to fail")
	}
}
