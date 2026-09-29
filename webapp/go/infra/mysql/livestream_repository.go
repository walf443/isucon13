package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type livestreamRepository struct{}

func NewLivestreamRepository() repository.LivestreamRepository {
	return &livestreamRepository{}
}

func (r *livestreamRepository) FindByID(ctx context.Context, q repository.Querier, id domain.LivestreamID) (*domain.Livestream, error) {
	var livestreamModel domain.Livestream
	err := q.GetContext(ctx, &livestreamModel, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &livestreamModel, nil
}

func (r *livestreamRepository) FindAllByIDAndUserID(ctx context.Context, q repository.Querier, id domain.LivestreamID, userID domain.UserID) ([]*domain.Livestream, error) {
	var livestreamModels []*domain.Livestream
	if err := q.SelectContext(ctx, &livestreamModels, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE id = ? AND user_id = ?", id, userID); err != nil {
		return nil, err
	}
	return livestreamModels, nil
}

func (r *livestreamRepository) FindAll(ctx context.Context, q repository.Querier) ([]*domain.Livestream, error) {
	var livestreamModels []*domain.Livestream
	if err := q.SelectContext(ctx, &livestreamModels, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams"); err != nil {
		return nil, err
	}
	return livestreamModels, nil
}

func (r *livestreamRepository) FindAllByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) ([]*domain.Livestream, error) {
	var livestreamModels []*domain.Livestream
	if err := q.SelectContext(ctx, &livestreamModels, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE user_id = ?", userID); err != nil {
		return nil, err
	}
	return livestreamModels, nil
}

func (r *livestreamRepository) FindAllOrderByIDDesc(ctx context.Context, q repository.Querier) ([]*domain.Livestream, error) {
	var livestreamModels []*domain.Livestream
	if err := q.SelectContext(ctx, &livestreamModels, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams ORDER BY id DESC"); err != nil {
		return nil, err
	}
	return livestreamModels, nil
}

func (r *livestreamRepository) FindAllOrderByIDDescLimited(ctx context.Context, q repository.Querier, limit domain.Limit) ([]*domain.Livestream, error) {
	var livestreamModels []*domain.Livestream
	if err := q.SelectContext(ctx, &livestreamModels, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams ORDER BY id DESC LIMIT ?", limit); err != nil {
		return nil, err
	}
	return livestreamModels, nil
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
