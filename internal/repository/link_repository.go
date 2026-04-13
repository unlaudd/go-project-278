// Package repository provides data access layer for link entities.
// It abstracts database operations using sqlc-generated queries.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"url-shortener/internal/db"
)

// Link represents a shortened URL entity returned by the API.
// Fields are tagged for JSON serialization.
type Link struct {
	ID          int32  `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortURL    string `json:"short_url"` // Constructed from baseURL + "/r/" + ShortName
}

// LinkRepository defines the interface for link data operations.
// All methods accept context for cancellation and timeout support.
type LinkRepository interface {
	Create(ctx context.Context, link *Link, baseURL string) error
	GetByID(ctx context.Context, id int32, baseURL string) (*Link, error)
	GetByShortName(ctx context.Context, shortName string, baseURL string) (*Link, error)
	List(ctx context.Context, limit, offset int32, baseURL string) ([]*Link, error)
	Update(ctx context.Context, id int32, originalURL, shortName *string, baseURL string) (*Link, error)
	Delete(ctx context.Context, id int32) error
	Count(ctx context.Context) (int64, error)
}

type linkRepository struct {
	queries *db.Queries // sqlc-generated query methods
	db      *sql.DB     // Raw connection for transactions if needed
}

// NewLinkRepository creates a new LinkRepository instance.
// dbConn should be an already-initialized and pinged *sql.DB connection.
func NewLinkRepository(dbConn *sql.DB) LinkRepository {
	return &linkRepository{
		queries: db.New(dbConn),
		db:      dbConn,
	}
}

// toLink converts a sqlc-generated db.Link model to the public Link struct.
// It constructs the ShortURL field using the provided baseURL.
// Note: sqlc converts snake_case (original_url) to PascalCase (OriginalUrl).
func toLink(d db.Link, baseURL string) *Link {
	return &Link{
		ID:          d.ID,
		OriginalURL: d.OriginalUrl,
		ShortName:   d.ShortName,
		ShortURL:    baseURL + "/r/" + d.ShortName,
	}
}

// Create inserts a new link into the database.
// On success, it populates link.ID and link.ShortURL with generated values.
// Returns a sentinel error "short_name already exists" on unique constraint violation.
func (r *linkRepository) Create(ctx context.Context, link *Link, baseURL string) error {
	res, err := r.queries.CreateLink(ctx, db.CreateLinkParams{
		OriginalUrl: link.OriginalURL,
		ShortName:   link.ShortName,
	})
	if err != nil {
		// Simple string matching for uniqueness error — consider using errors.As with pq.Error in production.
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("short_name already exists")
		}
		return err
	}
	link.ID = res.ID
	link.ShortURL = baseURL + "/r/" + res.ShortName
	return nil
}

// GetByID retrieves a link by its numeric ID.
// Returns "link not found" error if no record matches the ID.
func (r *linkRepository) GetByID(ctx context.Context, id int32, baseURL string) (*Link, error) {
	res, err := r.queries.GetLinkByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("link not found")
	}
	if err != nil {
		return nil, err
	}
	return toLink(res, baseURL), nil
}

// GetByShortName retrieves a link by its short identifier.
// Used for redirect resolution (GET /r/:shortName).
func (r *linkRepository) GetByShortName(ctx context.Context, shortName string, baseURL string) (*Link, error) {
	res, err := r.queries.GetLinkByShortName(ctx, shortName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("link not found")
	}
	if err != nil {
		return nil, err
	}
	return toLink(res, baseURL), nil
}

// List returns a paginated slice of links ordered by creation time (newest first).
// limit and offset control pagination; both are inclusive per API design.
func (r *linkRepository) List(ctx context.Context, limit, offset int32, baseURL string) ([]*Link, error) {
	rows, err := r.queries.ListLinks(ctx, db.ListLinksParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	links := make([]*Link, 0, len(rows))
	for _, row := range rows {
		links = append(links, toLink(row, baseURL))
	}
	return links, nil
}

// Update modifies an existing link's fields.
// Parameters originalURL and shortName are pointers: if nil, the existing value is preserved.
// This allows partial updates (PATCH semantics) via a single method.
// Returns "short_name already exists" if the new short_name conflicts with another link.
func (r *linkRepository) Update(ctx context.Context, id int32, originalURL, shortName *string, baseURL string) (*Link, error) {
	// Fetch current values to use as fallbacks for nil pointers.
	current, err := r.queries.GetLinkByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("link not found")
	}
	if err != nil {
		return nil, err
	}

	// Apply partial update logic: use new value if provided, otherwise keep existing.
	updOriginalURL := current.OriginalUrl
	if originalURL != nil {
		updOriginalURL = *originalURL
	}

	updShortName := current.ShortName
	if shortName != nil {
		updShortName = *shortName
	}

	params := db.UpdateLinkParams{
		ID:          id,
		OriginalUrl: updOriginalURL,
		ShortName:   updShortName,
	}

	res, err := r.queries.UpdateLink(ctx, params)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, errors.New("short_name already exists")
		}
		return nil, err
	}
	return toLink(res, baseURL), nil
}

// Delete removes a link by ID.
// Returns "link not found" if the ID does not exist.
// Note: Due to ON DELETE CASCADE, associated link_visits are automatically removed.
func (r *linkRepository) Delete(ctx context.Context, id int32) error {
	// Check existence before deletion to return a meaningful error.
	_, err := r.queries.GetLinkByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("link not found")
	}
	if err != nil {
		return err
	}
	return r.queries.DeleteLink(ctx, id)
}

// Count returns the total number of links in the database.
// Used for pagination metadata (Content-Range header).
func (r *linkRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountLinks(ctx)
}
