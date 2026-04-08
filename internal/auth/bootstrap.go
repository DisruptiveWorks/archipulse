package auth

import (
	"errors"
	"fmt"
)

// Bootstrap ensures the first admin user exists when the DB is empty.
// It is a no-op when users already exist or when the env vars are not set.
func Bootstrap(svc *Service) error {
	if svc.Cfg.BootstrapEmail == "" || svc.Cfg.BootstrapPassword == "" {
		return nil
	}

	exists, err := svc.Users.Exists()
	if err != nil {
		return fmt.Errorf("bootstrap check: %w", err)
	}
	if exists {
		return nil
	}

	hash, err := HashPassword(svc.Cfg.BootstrapPassword)
	if err != nil {
		return fmt.Errorf("bootstrap hash password: %w", err)
	}

	_, err = svc.Users.Create(svc.Cfg.BootstrapEmail, hash, "admin")
	if err != nil {
		return fmt.Errorf("bootstrap create admin: %w", err)
	}

	fmt.Printf("auth: bootstrapped admin user %q\n", svc.Cfg.BootstrapEmail)
	return nil
}

// LoginLocal authenticates a user by email + password.
// Returns the signed JWT string on success.
func LoginLocal(svc *Service, email, password string) (string, error) {
	u, err := svc.Users.GetByEmail(email)
	if errors.Is(err, ErrNotFound) {
		return "", fmt.Errorf("invalid credentials")
	}
	if err != nil {
		return "", err
	}
	if u.PasswordHash == nil || !CheckPassword(*u.PasswordHash, password) {
		return "", fmt.Errorf("invalid credentials")
	}
	return IssueToken(svc.Cfg, u.ID.String(), u.Email, u.Role)
}
