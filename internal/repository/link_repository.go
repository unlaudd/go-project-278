package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"url-shortener/internal/db"
)

type Link struct {
	ID          int32  `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortURL    string `json:"short_url"`
}

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
	queries *db.Queries
	db      *sql.DB
}

func NewLinkRepository(dbConn *sql.DB) LinkRepository {
	return &linkRepository{
		queries: db.New(dbConn), // sqlc генерирует конструктор db.New()
		db:      dbConn,
	}
}

// Вспомогательная функция маппинга из sqlc-модели в нашу
func toLink(d db.Link, baseURL string) *Link {
	return &Link{
		ID:          d.ID,
		OriginalURL: d.OriginalUrl, // sqlc преобразует original_url -> OriginalUrl
		ShortName:   d.ShortName,
		ShortURL:    baseURL + "/r/" + d.ShortName,
	}
}

func (r *linkRepository) Create(ctx context.Context, link *Link, baseURL string) error {
	res, err := r.queries.CreateLink(ctx, db.CreateLinkParams{
		OriginalUrl: link.OriginalURL,
		ShortName:   link.ShortName,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("short_name already exists")
		}
		return err
	}
	link.ID = res.ID
	link.ShortURL = baseURL + "/r/" + res.ShortName
	return nil
}

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

func (r *linkRepository) Update(ctx context.Context, id int32, originalURL, shortName *string, baseURL string) (*Link, error) {
	// Получаем текущие значения, чтобы подставить их, если пришли nil
	current, err := r.queries.GetLinkByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("link not found")
	}
	if err != nil {
		return nil, err
	}

	// Если указатель nil → оставляем старое значение, иначе берём новое
	updOriginalURL := current.OriginalUrl
	if originalURL != nil {
		updOriginalURL = *originalURL
	}

	updShortName := current.ShortName
	if shortName != nil {
		updShortName = *shortName
	}

	// Передаём обычные string, как ожидает sqlc
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

func (r *linkRepository) Delete(ctx context.Context, id int32) error {
	_, err := r.queries.GetLinkByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("link not found")
	}
	if err != nil {
		return err
	}
	return r.queries.DeleteLink(ctx, id)
}

// Count возвращает общее количество ссылок в БД
func (r *linkRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountLinks(ctx)
}
