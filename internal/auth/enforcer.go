package auth

import (
	"database/sql"
	_ "embed"
	"fmt"

	casbinv2 "github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

//go:embed rbac_model.conf
var rbacModelConf string

// Enforcer wraps the Casbin enforcer and seed policy loading.
type Enforcer struct {
	e *casbinv2.Enforcer
}

// NewEnforcer creates and seeds a Casbin enforcer backed by PostgreSQL.
func NewEnforcer(db *sql.DB, cfg *Config) (*Enforcer, error) {
	m, err := model.NewModelFromString(rbacModelConf)
	if err != nil {
		return nil, fmt.Errorf("casbin model: %w", err)
	}

	adapter := newDBAdapter(db)
	e, err := casbinv2.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("casbin enforcer: %w", err)
	}

	// Load (or refresh) the policy from the DB.
	if err := e.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("casbin load policy: %w", err)
	}

	// Ensure base role-hierarchy and default policies exist.
	if err := seedPolicy(e); err != nil {
		return nil, fmt.Errorf("casbin seed policy: %w", err)
	}

	return &Enforcer{e: e}, nil
}

// seedPolicy writes the default role hierarchy and resource policies.
// It is idempotent — Casbin skips duplicates automatically.
func seedPolicy(e *casbinv2.Enforcer) error {
	// Role hierarchy: admin > architect > viewer
	roleHierarchy := [][2]string{
		{"admin", "architect"},
		{"architect", "viewer"},
	}
	for _, r := range roleHierarchy {
		if _, err := e.AddGroupingPolicy(r[0], r[1]); err != nil {
			return err
		}
	}

	// Policies: (role, resource_glob, action)
	policies := [][3]string{
		// admin can do everything
		{"admin", "/api/v1/*", "*"},

		// architect: full read+write on workspace resources, no user mgmt
		{"architect", "/api/v1/workspaces", "GET"},
		{"architect", "/api/v1/workspaces/*", "*"},

		// viewer: read-only on workspaces
		{"viewer", "/api/v1/workspaces", "GET"},
		{"viewer", "/api/v1/workspaces/*", "GET"},
	}
	for _, p := range policies {
		if _, err := e.AddPolicy(p[0], p[1], p[2]); err != nil {
			return err
		}
	}

	return e.SavePolicy()
}

// Allow reports whether the given role may perform act on obj.
func (en *Enforcer) Allow(role, obj, act string) (bool, error) {
	return en.e.Enforce(role, obj, act)
}
