package auth

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

// dbAdapter is a Casbin persist.Adapter backed by PostgreSQL using database/sql.
// It stores grouping rules (g) with three values: user, role, domain.
// Policy rules (p) are not used; enforcement logic lives in enforcer.go.
type dbAdapter struct {
	db *sql.DB
}

// LoadPolicy loads all casbin_rule rows into the model.
func (a *dbAdapter) LoadPolicy(m model.Model) error {
	rows, err := a.db.Query(
		"SELECT ptype, v0, v1, v2, v3, v4, v5 FROM casbin_rule",
	)
	if err != nil {
		return fmt.Errorf("load casbin policy: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var ptype, v0, v1, v2, v3, v4, v5 sql.NullString
		if err := rows.Scan(&ptype, &v0, &v1, &v2, &v3, &v4, &v5); err != nil {
			return err
		}
		parts := []string{ptype.String, v0.String, v1.String, v2.String, v3.String, v4.String, v5.String}
		// Trim trailing empty strings.
		n := len(parts)
		for n > 0 && parts[n-1] == "" {
			n--
		}
		if err := persist.LoadPolicyLine(strings.Join(parts[:n], ", "), m); err != nil {
			return err
		}
	}
	return rows.Err()
}

// SavePolicy replaces all policy rows atomically.
func (a *dbAdapter) SavePolicy(m model.Model) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec("DELETE FROM casbin_rule"); err != nil {
		return err
	}

	for ptype, assertions := range m {
		for _, assertion := range assertions {
			for _, rule := range assertion.Policy {
				row := append([]string{ptype}, rule...)
				if err := insertRule(tx, row); err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit()
}

func insertRule(tx *sql.Tx, parts []string) error {
	// Pad to 7 columns (ptype + v0..v5).
	for len(parts) < 7 {
		parts = append(parts, "")
	}
	_, err := tx.Exec(
		`INSERT INTO casbin_rule (ptype, v0, v1, v2, v3, v4, v5)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT DO NOTHING`,
		parts[0], parts[1], parts[2], parts[3], parts[4], parts[5], parts[6],
	)
	return err
}

func (a *dbAdapter) AddPolicy(sec, ptype string, rule []string) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := insertRule(tx, append([]string{ptype}, rule...)); err != nil {
		return err
	}
	return tx.Commit()
}

func (a *dbAdapter) RemovePolicy(sec, ptype string, rule []string) error {
	if len(rule) < 3 {
		return nil
	}
	// For grouping rules: v0=user, v1=role, v2=domain.
	_, err := a.db.Exec(
		`DELETE FROM casbin_rule WHERE ptype = $1 AND v0 = $2 AND v1 = $3 AND v2 = $4`,
		ptype, rule[0], rule[1], rule[2],
	)
	return err
}

func (a *dbAdapter) RemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	if len(fieldValues) == 0 {
		return nil
	}
	cols := []string{"v0", "v1", "v2", "v3", "v4", "v5"}
	conds := []string{"ptype = $1"}
	args := []any{ptype}
	for i, v := range fieldValues {
		if v == "" {
			continue
		}
		col := cols[fieldIndex+i]
		args = append(args, v)
		conds = append(conds, fmt.Sprintf("%s = $%d", col, len(args)))
	}
	q := "DELETE FROM casbin_rule WHERE " + strings.Join(conds, " AND ")
	_, err := a.db.Exec(q, args...)
	return err
}

// Ensure interface compliance.
var _ persist.Adapter = (*dbAdapter)(nil)
