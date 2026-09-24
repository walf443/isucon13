package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type livestreamRepository struct {
	// fallbackIcon はアイコン未登録の配信者に使う画像。
	fallbackIcon []byte
}

func NewLivestreamRepository(fallbackIcon []byte) repository.LivestreamRepository {
	return &livestreamRepository{fallbackIcon: fallbackIcon}
}

func (r *livestreamRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id int64) (*model.Livestream, error) {
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
