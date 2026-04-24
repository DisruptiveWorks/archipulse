package views

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// DomainApp is an application entry in the domain landscape.
type DomainApp struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	Domain     string            `json:"domain"`
	Properties map[string]string `json:"properties"`
}

// DomainGroup is a named domain with its applications.
type DomainGroup struct {
	ID   string      `json:"id"`
	Name string      `json:"name"`
	Apps []DomainApp `json:"apps"`
}

// DomainLandscapeData is the payload for the Application Landscape (by domain) view.
type DomainLandscapeData struct {
	Domains    []DomainGroup `json:"domains"`
	Properties []string      `json:"properties"`
}

// ApplicationLandscapeDomain groups ApplicationComponent elements by their
// 'domain' property (or 'business_domain'). Apps without either property are
// placed in an "Uncategorized" group.
func ApplicationLandscapeDomain(db *sql.DB, workspaceID uuid.UUID) (*DomainLandscapeData, error) {
	// 1. Fetch all application elements with domain property.
	rows, err := db.Query(`
		SELECT
			e.source_id,
			e.name,
			e.type,
			COALESCE(
				(SELECT ep.value FROM element_properties ep
				 WHERE ep.element_id = e.id AND ep.key IN ('domain','business_domain')
				 ORDER BY CASE ep.key WHEN 'domain' THEN 0 ELSE 1 END LIMIT 1),
				''
			) AS domain
		FROM elements e
		WHERE e.workspace_id = $1
		  AND e.type IN (`+appTypesSQL+`)
		ORDER BY domain, e.name`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("domain landscape apps: %w", err)
	}
	defer func() { _ = rows.Close() }()

	type appRow struct {
		id     string
		name   string
		typ    string
		domain string
	}
	var appRows []appRow
	var appIDs []string

	for rows.Next() {
		var r appRow
		if err := rows.Scan(&r.id, &r.name, &r.typ, &r.domain); err != nil {
			return nil, err
		}
		appRows = append(appRows, r)
		appIDs = append(appIDs, r.id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 2. Load properties for all apps.
	propsByApp, propKeys, err := loadAppProperties(db, workspaceID, appIDs)
	if err != nil {
		return nil, err
	}

	// 3. Group by domain.
	domainMap := map[string]*DomainGroup{}
	domainOrder := []string{}

	for _, ar := range appRows {
		domain := ar.domain
		if domain == "" {
			domain = "Uncategorized"
		}
		if _, exists := domainMap[domain]; !exists {
			domainMap[domain] = &DomainGroup{ID: domain, Name: domain, Apps: []DomainApp{}}
			domainOrder = append(domainOrder, domain)
		}
		app := DomainApp{
			ID:         ar.id,
			Name:       ar.name,
			Type:       ar.typ,
			Domain:     domain,
			Properties: map[string]string{},
		}
		if p, ok := propsByApp[ar.id]; ok {
			app.Properties = p
		}
		domainMap[domain].Apps = append(domainMap[domain].Apps, app)
	}

	// Sort domains alphabetically, Uncategorized last.
	sort.Slice(domainOrder, func(i, j int) bool {
		a, b := domainOrder[i], domainOrder[j]
		if a == "Uncategorized" {
			return false
		}
		if b == "Uncategorized" {
			return true
		}
		return a < b
	})

	domains := make([]DomainGroup, 0, len(domainOrder))
	for _, d := range domainOrder {
		domains = append(domains, *domainMap[d])
	}

	sort.Strings(propKeys)
	return &DomainLandscapeData{Domains: domains, Properties: propKeys}, nil
}
