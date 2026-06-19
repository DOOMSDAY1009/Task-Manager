package auth

import (
	"testing"
	"time"

	"github.com/hireft/task-manager/internal/models"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("s3cret-password", 4) // low cost keeps the test fast
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if hash == "s3cret-password" {
		t.Fatal("password was not hashed")
	}
	if !CheckPassword(hash, "s3cret-password") {
		t.Error("correct password should verify")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Error("wrong password should not verify")
	}
}

func TestJWTGenerateAndParse(t *testing.T) {
	mgr := NewManager("this-is-a-test-secret-key", time.Hour)
	user := models.User{ID: "user-123", Email: "a@b.com", Role: models.RoleAdmin}

	token, err := mgr.Generate(user, time.Now())
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	claims, err := mgr.Parse(token)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want user-123", claims.UserID)
	}
	if claims.Role != models.RoleAdmin {
		t.Errorf("Role = %q, want admin", claims.Role)
	}
}

func TestJWTRejectsExpiredToken(t *testing.T) {
	mgr := NewManager("this-is-a-test-secret-key", time.Hour)
	user := models.User{ID: "u1", Email: "a@b.com", Role: models.RoleUser}

	// Issue a token that expired two hours ago.
	token, err := mgr.Generate(user, time.Now().Add(-2*time.Hour))
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if _, err := mgr.Parse(token); err == nil {
		t.Error("expected expired token to be rejected")
	}
}

func TestJWTRejectsWrongSecret(t *testing.T) {
	issuer := NewManager("the-original-secret-key", time.Hour)
	attacker := NewManager("a-totally-different-key", time.Hour)
	token, _ := issuer.Generate(models.User{ID: "u1"}, time.Now())

	if _, err := attacker.Parse(token); err == nil {
		t.Error("token signed with a different secret should be rejected")
	}
}
