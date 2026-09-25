package mysql

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type livecommentRepository struct {
	// fallbackIcon はアイコン未登録のユーザに使う画像。
	fallbackIcon []byte
}

func NewLivecommentRepository(fallbackIcon []byte) repository.LivecommentRepository {
	return &livecommentRepository{fallbackIcon: fallbackIcon}
}

func (r *livecommentRepository) FindAllWithDetailsByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.Livecomment, error) {
	var livecommentModels []*model.LivecommentModel
	if err := q.SelectContext(ctx, &livecommentModels, "SELECT id, user_id, livestream_id, comment, tip, created_at FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC", livestreamID); err != nil {
		return nil, err
	}
	return fillLivecomments(ctx, q, livecommentModels, r.fallbackIcon)
}

func (r *livecommentRepository) FindAllWithDetailsByLivestreamIDLimited(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID, limit int64) ([]*model.Livecomment, error) {
	var livecommentModels []*model.LivecommentModel
	if err := q.SelectContext(ctx, &livecommentModels, "SELECT id, user_id, livestream_id, comment, tip, created_at FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC LIMIT ?", livestreamID, limit); err != nil {
		return nil, err
	}
	return fillLivecomments(ctx, q, livecommentModels, r.fallbackIcon)
}
