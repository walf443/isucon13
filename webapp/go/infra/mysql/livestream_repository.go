package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
	"github.com/jmoiron/sqlx"
)

type livestreamRepository struct {
	// defaultIconHash はアイコン未登録の配信者に使う既定のアイコンのハッシュ。
	defaultIconHash model.IconHash
}

func NewLivestreamRepository(defaultIconHash model.IconHash) repository.LivestreamRepository {
	return &livestreamRepository{defaultIconHash: defaultIconHash}
}

func (r *livestreamRepository) FindByID(ctx context.Context, q repository.Querier, id model.LivestreamID) (*model.LivestreamModel, error) {
	var livestreamModel model.LivestreamModel
	err := q.GetContext(ctx, &livestreamModel, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &livestreamModel, nil
}

func (r *livestreamRepository) FindAllByIDAndUserID(ctx context.Context, q repository.Querier, id model.LivestreamID, userID model.UserID) ([]*model.LivestreamModel, error) {
	var livestreamModels []*model.LivestreamModel
	if err := q.SelectContext(ctx, &livestreamModels, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE id = ? AND user_id = ?", id, userID); err != nil {
		return nil, err
	}
	return livestreamModels, nil
}

func (r *livestreamRepository) FindAll(ctx context.Context, q repository.Querier) ([]*model.LivestreamModel, error) {
	var livestreamModels []*model.LivestreamModel
	if err := q.SelectContext(ctx, &livestreamModels, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams"); err != nil {
		return nil, err
	}
	return livestreamModels, nil
}

func (r *livestreamRepository) FindAllByUserID(ctx context.Context, q repository.Querier, userID model.UserID) ([]*model.LivestreamModel, error) {
	var livestreamModels []*model.LivestreamModel
	if err := q.SelectContext(ctx, &livestreamModels, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE user_id = ?", userID); err != nil {
		return nil, err
	}
	return livestreamModels, nil
}

func (r *livestreamRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id model.LivestreamID) (*model.Livestream, error) {
	var livestreamModel model.LivestreamModel
	err := q.GetContext(ctx, &livestreamModel, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	livestreams, err := fillLivestreams(ctx, q, []*model.LivestreamModel{&livestreamModel}, r.defaultIconHash)
	if err != nil {
		return nil, err
	}
	return livestreams[0], nil
}

func (r *livestreamRepository) FindAllWithDetailsByUserID(ctx context.Context, q repository.Querier, userID model.UserID) ([]*model.Livestream, error) {
	var livestreamModels []*model.LivestreamModel
	if err := q.SelectContext(ctx, &livestreamModels, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE user_id = ?", userID); err != nil {
		return nil, err
	}
	return fillLivestreams(ctx, q, livestreamModels, r.defaultIconHash)
}

func (r *livestreamRepository) FindAllWithDetailsByTagIDs(ctx context.Context, q repository.Querier, tagIDs []model.TagID) ([]*model.Livestream, error) {
	query, params, err := sqlx.In("SELECT id, livestream_id, tag_id FROM livestream_tags WHERE tag_id IN (?) ORDER BY livestream_id DESC", tagIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to construct IN query: %w", err)
	}
	var livestreamTagModels []*livestreamTagModel
	if err := q.SelectContext(ctx, &livestreamTagModels, query, params...); err != nil {
		return nil, err
	}

	livestreamModels := make([]*model.LivestreamModel, len(livestreamTagModels))
	for i, livestreamTagModel := range livestreamTagModels {
		var livestreamModel model.LivestreamModel
		if err := q.GetContext(ctx, &livestreamModel, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE id = ?", livestreamTagModel.LivestreamID); err != nil {
			return nil, fmt.Errorf("failed to get livestream %d: %w", livestreamTagModel.LivestreamID, err)
		}
		livestreamModels[i] = &livestreamModel
	}
	return fillLivestreams(ctx, q, livestreamModels, r.defaultIconHash)
}

func (r *livestreamRepository) FindAllWithDetails(ctx context.Context, q repository.Querier) ([]*model.Livestream, error) {
	var livestreamModels []*model.LivestreamModel
	if err := q.SelectContext(ctx, &livestreamModels, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams ORDER BY id DESC"); err != nil {
		return nil, err
	}
	return fillLivestreams(ctx, q, livestreamModels, r.defaultIconHash)
}

func (r *livestreamRepository) FindAllWithDetailsLimited(ctx context.Context, q repository.Querier, limit model.Limit) ([]*model.Livestream, error) {
	var livestreamModels []*model.LivestreamModel
	if err := q.SelectContext(ctx, &livestreamModels, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams ORDER BY id DESC LIMIT ?", limit); err != nil {
		return nil, err
	}
	return fillLivestreams(ctx, q, livestreamModels, r.defaultIconHash)
}

func (r *livestreamRepository) Create(ctx context.Context, q repository.Querier, livestream *model.LivestreamModel) (model.LivestreamID, error) {
	rs, err := q.ExecContext(ctx, "INSERT INTO livestreams (user_id, title, description, playlist_url, thumbnail_url, start_at, end_at) VALUES(?, ?, ?, ?, ?, ?, ?)", livestream.UserID, livestream.Title, livestream.Description, livestream.PlaylistUrl, livestream.ThumbnailUrl, livestream.StartAt, livestream.EndAt)
	if err != nil {
		return 0, err
	}
	id, err := rs.LastInsertId()
	if err != nil {
		return 0, err
	}
	return model.LivestreamID(id), nil
}

func (r *livestreamRepository) AddTag(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID, tagID model.TagID) error {
	_, err := q.ExecContext(ctx, "INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (?, ?)", livestreamID, tagID)
	return err
}
