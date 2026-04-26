// Package pagination provides shared types for paginated API list endpoints.
package pagination

const defaultLimit = 100
const maxLimit = 500

// Params holds normalized pagination inputs derived from query parameters.
type Params struct {
	Page  int
	Limit int
}

// Normalize returns a Params with sane defaults and bounds applied.
func Normalize(page, limit int) Params {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return Params{Page: page, Limit: limit}
}

// Offset returns the SQL OFFSET value for this page.
func (p Params) Offset() int { return (p.Page - 1) * p.Limit }

// Page is the generic paginated response envelope.
type Page[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// NewPage wraps items in a Page, ensuring Items is never null in JSON.
func NewPage[T any](items []T, total, page, limit int) Page[T] {
	if items == nil {
		items = []T{}
	}
	return Page[T]{Items: items, Total: total, Page: page, Limit: limit}
}
