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
	// fallbackIcon はアイコン未登録の配信者に使う画像。
	fallbackIcon []byte
}

func NewLivestreamRepository(fallbackIcon []byte) repository.LivestreamRepository {
	return &livestreamRepository{fallbackIcon: fallbackIcon}
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

func (r *livestreamRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id model.LivestreamID) (*model.Livestream, error) {
	var livestreamModel model.LivestreamModel
	err := q.GetContext(ctx, &livestreamModel, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	livestreams, err := fillLivestreams(ctx, q, []*model.LivestreamModel{&livestreamModel}, r.fallbackIcon)
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
	return fillLivestreams(ctx, q, livestreamModels, r.fallbackIcon)
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
	return fillLivestreams(ctx, q, livestreamModels, r.fallbackIcon)
}

func (r *livestreamRepository) FindAllWithDetails(ctx context.Context, q repository.Querier) ([]*model.Livestream, error) {
	var livestreamModels []*model.LivestreamModel
	if err := q.SelectContext(ctx, &livestreamModels, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams ORDER BY id DESC"); err != nil {
		return nil, err
	}
	return fillLivestreams(ctx, q, livestreamModels, r.fallbackIcon)
}

func (r *livestreamRepository) FindAllWithDetailsLimited(ctx context.Context, q repository.Querier, limit int64) ([]*model.Livestream, error) {
	var livestreamModels []*model.LivestreamModel
	if err := q.SelectContext(ctx, &livestreamModels, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams ORDER BY id DESC LIMIT ?", limit); err != nil {
		return nil, err
	}
	return fillLivestreams(ctx, q, livestreamModels, r.fallbackIcon)
}
