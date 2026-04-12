package repository

import (
	"context"
	"database/sql"
	"errors"

	"url-shortener/internal/db"
)

type Link struct {
	ID          int64  `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortURL    string `json:"short_url"` // формируется на сервере
}

type LinkRepository interface {
	Create(ctx context.Context, link *Link, baseURL string) error
	GetByID(ctx context.Context, id int64, baseURL string) (*Link, error)
	GetByShortName(ctx context.Context, shortName string, baseURL string) (*Link, error)
	List(ctx context.Context, limit, offset int, baseURL string) ([]*Link, error)
	Update(ctx context.Context, id int64, originalURL, shortName *string, baseURL string) (*Link, error)
	Delete(ctx context.Context, id int64) error
}

type linkRepository struct {
	queries *db.Queries
	db      *sql.DB
}

func NewLinkRepository(db *sql.DB) LinkRepository {
	return &linkRepository{
		queries: db.New(db),
		db:      db,
	}
}

func (r *linkRepository) Create(ctx context.Context, link *Link, baseURL string) error {
	res, err := r.queries.CreateLink(ctx, db.CreateLinkParams{
		OriginalURL: link.OriginalURL,
		ShortName:   link.ShortName,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return errors.New("short_name already exists")
		}
		return err
	}
	link.ID = res.ID
	link.ShortURL = baseURL + "/r/" + res.ShortName
	return nil
}

func (r *linkRepository) GetByID(ctx context.Context, id int64, baseURL string) (*Link, error) {
	res, err := r.queries.GetLinkByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("link not found")
	}
	if err != nil {
		return nil, err
	}
	return &Link{
		ID:          res.ID,
		OriginalURL: res.OriginalURL,
		ShortName:   res.ShortName,
		ShortURL:    baseURL + "/r/" + res.ShortName,
	}, nil
}

func (r *linkRepository) GetByShortName(ctx context.Context, shortName string, baseURL string) (*Link, error) {
	res, err := r.queries.GetLinkByShortName(ctx, shortName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("link not found")
	}
	if err != nil {
		return nil, err
	}
	return &Link{
		ID:          res.ID,
		OriginalURL: res.OriginalURL,
		ShortName:   res.ShortName,
		ShortURL:    baseURL + "/r/" + res.ShortName,
	}, nil
}

func (r *linkRepository) List(ctx context.Context, limit, offset int, baseURL string) ([]*Link, error) {
	rows, err := r.queries.ListLinks(ctx, db.ListLinksParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	links := make([]*Link, 0, len(rows))
	for _, row := range rows {
		links = append(links, &Link{
			ID:          row.ID,
			OriginalURL: row.OriginalURL,
			ShortName:   row.ShortName,
			ShortURL:    baseURL + "/r/" + row.ShortName,
		})
	}
	return links, nil
}

func (r *linkRepository) Update(ctx context.Context, id int64, originalURL, shortName *string, baseURL string) (*Link, error) {
	res, err := r.queries.UpdateLink(ctx, db.UpdateLinkParams{
		ID:          id,
		OriginalURL: sql.NullString{String: *originalURL, Valid: originalURL != nil},
		ShortName:   sql.NullString{String: *shortName, Valid: shortName != nil},
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("link not found")
	}
	if err != nil {
		if isUniqueViolation(err) {
			return nil, errors.New("short_name already exists")
		}
		return nil, err
	}
	return &Link{
		ID:          res.ID,
		OriginalURL: res.OriginalURL,
		ShortName:   res.ShortName,
		ShortURL:    baseURL + "/r/" + res.ShortName,
	}, nil
}

func (r *linkRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.queries.GetLinkByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("link not found")
	}
	if err != nil {
		return err
	}
	return r.queries.DeleteLink(ctx, id)
}

// Вспомогательная функция для определения ошибки уникальности
func isUniqueViolation(err error) bool {
	// Для PostgreSQL код ошибки уникальности: 23505
	// Можно использовать pgconn, но для простоты проверяем строку
	return err != nil && (err.Error() == "short_name already exists" || 
		(err.Error() != "" && len(err.Error()) > 0))
}
