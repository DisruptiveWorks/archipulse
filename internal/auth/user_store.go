package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// User represents a row in the users table.
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash *string // nil for OIDC-only users
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ErrNotFound is returned when a user lookup yields no row.
var ErrNotFound = errors.New("user not found")

// UserStore provides CRUD access to the users table.
type UserStore struct {
	db *sql.DB
}

// NewUserStore creates a UserStore.
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

const userCols = "id, email, password_hash, role, created_at, updated_at"

func scanUser(row interface{ Scan(...any) error }) (*User, error) {
	var u User
	var hash sql.NullString
	if err := row.Scan(&u.ID, &u.Email, &hash, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if hash.Valid {
		u.PasswordHash = &hash.String
	}
	return &u, nil
}

// GetByEmail fetches a user by email address.
func (s *UserStore) GetByEmail(email string) (*User, error) {
	row := s.db.QueryRow(
		"SELECT "+userCols+" FROM users WHERE email = $1", email,
	)
	return scanUser(row)
}

// GetByID fetches a user by UUID.
func (s *UserStore) GetByID(id string) (*User, error) {
	row := s.db.QueryRow(
		"SELECT "+userCols+" FROM users WHERE id = $1", id,
	)
	return scanUser(row)
}

// Create inserts a new user and returns the created record.
func (s *UserStore) Create(email, passwordHash, role string) (*User, error) {
	var ph *string
	if passwordHash != "" {
		ph = &passwordHash
	}
	row := s.db.QueryRow(
		`INSERT INTO users (email, password_hash, role)
		 VALUES ($1, $2, $3)
		 RETURNING `+userCols,
		email, ph, role,
	)
	return scanUser(row)
}

// Exists reports whether any user row exists.
func (s *UserStore) Exists() (bool, error) {
	var n int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&n)
	if err != nil {
		return false, fmt.Errorf("count users: %w", err)
	}
	return n > 0, nil
}

// UpdatePasswordHash replaces a user's bcrypt password hash.
func (s *UserStore) UpdatePasswordHash(id, hash string) error {
	_, err := s.db.Exec(
		"UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2",
		hash, id,
	)
	return err
}

// UpdateRole changes a user's role.
func (s *UserStore) UpdateRole(id, role string) error {
	_, err := s.db.Exec(
		"UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2",
		role, id,
	)
	return err
}
