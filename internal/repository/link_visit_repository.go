// Package repository provides data access layer for link visit analytics.
// It abstracts database operations using sqlc-generated queries.
package repository

import (
	"context"
	"database/sql"
	"time"

	"url-shortener/internal/db"
)

// LinkVisit represents a single redirect event with client metadata.
// Fields are tagged for JSON serialization in API responses.
type LinkVisit struct {
	ID        int32  `json:"id"`
	LinkID    int32  `json:"link_id"`    // Foreign key to the links table
	IP        string `json:"ip"`         // Client IP address (IPv4 or IPv6)
	UserAgent string `json:"user_agent"` // Optional browser/client identifier
	Referer   string `json:"referer"`    // Optional HTTP Referer header
	Status    int16  `json:"status"`     // HTTP status code of the redirect (e.g., 301, 404)
	CreatedAt string `json:"created_at"` // RFC3339-formatted timestamp
}

// LinkVisitRepository defines the interface for link visit data operations.
// All methods accept context for cancellation and timeout support.
type LinkVisitRepository interface {
	Create(ctx context.Context, visit *LinkVisit) error
	List(ctx context.Context, limit, offset int32) ([]*LinkVisit, error)
	Count(ctx context.Context) (int64, error)
}

type linkVisitRepository struct {
	queries *db.Queries // sqlc-generated query methods
}

// NewLinkVisitRepository creates a new LinkVisitRepository instance.
// dbConn should be an already-initialized *sql.DB connection.
func NewLinkVisitRepository(dbConn *sql.DB) LinkVisitRepository {
	return &linkVisitRepository{queries: db.New(dbConn)}
}

// Create inserts a new visit record into the database.
// On success, it populates visit.ID and visit.CreatedAt with generated values.
// Note: Status is int16 to match PostgreSQL SMALLINT; sqlc enforces this type safety.
func (r *linkVisitRepository) Create(ctx context.Context, visit *LinkVisit) error {
	res, err := r.queries.CreateLinkVisit(ctx, db.CreateLinkVisitParams{
		LinkID: visit.LinkID,
		Ip:     visit.IP,
		// Use sql.NullString for optional TEXT fields to handle empty values correctly.
		UserAgent: sql.NullString{String: visit.UserAgent, Valid: visit.UserAgent != ""},
		Referer:   sql.NullString{String: visit.Referer, Valid: visit.Referer != ""},
		Status:    visit.Status,
	})
	if err != nil {
		return err
	}
	visit.ID = res.ID

	// Format timestamp for JSON output. sqlc returns sql.NullTime for nullable TIMESTAMPTZ.
	if res.CreatedAt.Valid {
		visit.CreatedAt = res.CreatedAt.Time.Format(time.RFC3339)
	} else {
		// Fallback: should not happen with DEFAULT NOW(), but guard against nil.
		visit.CreatedAt = time.Now().Format(time.RFC3339)
	}
	return nil
}

// List returns a paginated slice of visits ordered by creation time (newest first).
// limit and offset control pagination; both are inclusive per API design.
func (r *linkVisitRepository) List(ctx context.Context, limit, offset int32) ([]*LinkVisit, error) {
	rows, err := r.queries.ListLinkVisits(ctx, db.ListLinkVisitsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	visits := make([]*LinkVisit, 0, len(rows))
	for _, row := range rows {
		// Convert sql.NullTime to RFC3339 string for JSON serialization.
		createdAt := ""
		if row.CreatedAt.Valid {
			createdAt = row.CreatedAt.Time.Format(time.RFC3339)
		}
		visits = append(visits, &LinkVisit{
			ID:     row.ID,
			LinkID: row.LinkID,
			IP:     row.Ip,
			// Extract string value from sql.NullString; empty if not valid.
			UserAgent: row.UserAgent.String,
			Referer:   row.Referer.String,
			Status:    row.Status,
			CreatedAt: createdAt,
		})
	}
	return visits, nil
}

// Count returns the total number of visit records in the database.
// Used for pagination metadata (Content-Range header) in the API.
func (r *linkVisitRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountLinkVisits(ctx)
}
