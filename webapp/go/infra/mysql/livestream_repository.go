package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"github.com/jmoiron/sqlx"
)

type livestreamRepository struct{}

func NewLivestreamRepository() repository.LivestreamRepository {
	return &livestreamRepository{}
}

func (r *livestreamRepository) FindByID(ctx context.Context, q repository.Querier, id domain.LivestreamID) (*domain.Livestream, error) {
	var livestream domain.Livestream
	err := q.GetContext(ctx, &livestream, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &livestream, nil
}

func (r *livestreamRepository) FindAllByIDAndUserID(ctx context.Context, q repository.Querier, id domain.LivestreamID, userID domain.UserID) ([]*domain.Livestream, error) {
	var livestreams []*domain.Livestream
	if err := q.SelectContext(ctx, &livestreams, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE id = ? AND user_id = ?", id, userID); err != nil {
		return nil, err
	}
	return livestreams, nil
}

func (r *livestreamRepository) FindAll(ctx context.Context, q repository.Querier) ([]*domain.Livestream, error) {
	var livestreams []*domain.Livestream
	if err := q.SelectContext(ctx, &livestreams, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams"); err != nil {
		return nil, err
	}
	return livestreams, nil
}

func (r *livestreamRepository) FindAllByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) ([]*domain.Livestream, error) {
	var livestreams []*domain.Livestream
	if err := q.SelectContext(ctx, &livestreams, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE user_id = ?", userID); err != nil {
		return nil, err
	}
	return livestreams, nil
}

func (r *livestreamRepository) FindAllOrderByIDDesc(ctx context.Context, q repository.Querier) ([]*domain.Livestream, error) {
	var livestreams []*domain.Livestream
	if err := q.SelectContext(ctx, &livestreams, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams ORDER BY id DESC"); err != nil {
		return nil, err
	}
	return livestreams, nil
}

func (r *livestreamRepository) FindAllOrderByIDDescLimited(ctx context.Context, q repository.Querier, limit domain.Limit) ([]*domain.Livestream, error) {
	var livestreams []*domain.Livestream
	if err := q.SelectContext(ctx, &livestreams, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams ORDER BY id DESC LIMIT ?", limit); err != nil {
		return nil, err
	}
	return livestreams, nil
}

func (r *livestreamRepository) Create(ctx context.Context, q repository.Querier, livestream *domain.Livestream) (domain.LivestreamID, error) {
	rs, err := q.ExecContext(ctx, "INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES(?, ?, ?, ?, ?, ?, ?)", livestream.UserID, livestream.Title, livestream.Description, livestream.PlaylistUrl, livestream.ThumbnailUrl, livestream.StartAt, livestream.EndAt)
	if err != nil {
		return 0, err
	}
	id, err := rs.LastInsertId()
	if err != nil {
		return 0, err
	}
	return domain.LivestreamID(id), nil
}

func (r *livestreamRepository) FindAllByIDs(ctx context.Context, q repository.Querier, ids []domain.LivestreamID) ([]*domain.Livestream, error) {
	// IN () は作れないので、空の場合はクエリを発行しない
	if len(ids) == 0 {
		return []*domain.Livestream{}, nil
	}
	query, params, err := sqlx.In("SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE id IN (?)", ids)
	if err != nil {
		return nil, fmt.Errorf("failed to construct IN query: %w", err)
	}
	var livestreams []*domain.Livestream
	if err := q.SelectContext(ctx, &livestreams, query, params...); err != nil {
		return nil, err
	}
	return livestreams, nil
}
