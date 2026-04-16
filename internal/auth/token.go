package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the JWT payload stored in the ap_session cookie.
type Claims struct {
	UserID  string `json:"uid"`
	Email   string `json:"email"`
	OrgRole string `json:"org_role"`
	jwt.RegisteredClaims
}

// IssueToken signs a new JWT for the given user.
func IssueToken(cfg *Config, userID, email, orgRole string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:  userID,
		Email:   email,
		OrgRole: orgRole,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.TokenTTL)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(cfg.JWTSecret))
}

// ParseToken validates and parses a signed JWT string.
func ParseToken(cfg *Config, raw string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	c, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return c, nil
}
