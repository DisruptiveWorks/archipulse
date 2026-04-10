package auth_test

import (
	"testing"

	"github.com/DisruptiveWorks/archipulse/internal/auth"
)

func TestHashPassword_ProducesValidHash(t *testing.T) {
	hash, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if len(hash) < 50 {
		t.Errorf("hash too short: %q", hash)
	}
}

func TestHashPassword_DifferentSaltsEachCall(t *testing.T) {
	h1, _ := auth.HashPassword("same")
	h2, _ := auth.HashPassword("same")
	if h1 == h2 {
		t.Error("expected different hashes for the same password (different salts)")
	}
}

func TestCheckPassword_Correct(t *testing.T) {
	hash, _ := auth.HashPassword("mypassword")
	if !auth.CheckPassword(hash, "mypassword") {
		t.Error("CheckPassword returned false for correct password")
	}
}

func TestCheckPassword_Wrong(t *testing.T) {
	hash, _ := auth.HashPassword("mypassword")
	if auth.CheckPassword(hash, "wrongpassword") {
		t.Error("CheckPassword returned true for wrong password")
	}
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	hash, _ := auth.HashPassword("notempty")
	if auth.CheckPassword(hash, "") {
		t.Error("CheckPassword returned true for empty password")
	}
}
