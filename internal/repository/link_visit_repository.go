package repository

import (
	"context"
	"database/sql"
	"time"

	"url-shortener/internal/db"
)

type LinkVisit struct {
	ID        int32  `json:"id"`
	LinkID    int32  `json:"link_id"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Referer   string `json:"referer"`
	Status    int16  `json:"status"` // ← int16 для SMALLINT
	CreatedAt string `json:"created_at"`
}

type LinkVisitRepository interface {
	Create(ctx context.Context, visit *LinkVisit) error
	List(ctx context.Context, limit, offset int32) ([]*LinkVisit, error)
	Count(ctx context.Context) (int64, error)
}

type linkVisitRepository struct {
	queries *db.Queries
}

func NewLinkVisitRepository(dbConn *sql.DB) LinkVisitRepository {
	return &linkVisitRepository{queries: db.New(dbConn)}
}

func (r *linkVisitRepository) Create(ctx context.Context, visit *LinkVisit) error {
	res, err := r.queries.CreateLinkVisit(ctx, db.CreateLinkVisitParams{
		LinkID:    visit.LinkID,
		Ip:        visit.IP,
		UserAgent: sql.NullString{String: visit.UserAgent, Valid: visit.UserAgent != ""},
		Referer:   sql.NullString{String: visit.Referer, Valid: visit.Referer != ""},
		Status:    visit.Status, // int16
	})
	if err != nil {
		return err
	}
	visit.ID = res.ID
	// Обработка sql.NullTime
	if res.CreatedAt.Valid {
		visit.CreatedAt = res.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
	} else {
		visit.CreatedAt = time.Now().Format("2006-01-02T15:04:05Z")
	}
	return nil
}

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
		createdAt := ""
		if row.CreatedAt.Valid {
			createdAt = row.CreatedAt.Time.Format("2006-01-02T15:04:05Z")
		}
		visits = append(visits, &LinkVisit{
			ID:        row.ID,
			LinkID:    row.LinkID,
			IP:        row.Ip,
			UserAgent: row.UserAgent.String,
			Referer:   row.Referer.String,
			Status:    row.Status, // int16
			CreatedAt: createdAt,
		})
	}
	return visits, nil
}

func (r *linkVisitRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountLinkVisits(ctx)
}
