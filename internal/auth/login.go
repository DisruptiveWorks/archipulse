package auth

import "errors"

// LoginLocal authenticates a user by email + password.
// Returns the signed JWT string on success.
func LoginLocal(svc *Service, email, password string) (string, error) {
	u, err := svc.Users.GetByEmail(email)
	if errors.Is(err, ErrNotFound) {
		return "", errors.New("invalid credentials")
	}
	if err != nil {
		return "", err
	}
	if u.PasswordHash == nil || !CheckPassword(*u.PasswordHash, password) {
		return "", errors.New("invalid credentials")
	}
	return IssueToken(svc.Cfg, u.ID.String(), u.Email, u.OrgRole)
}
