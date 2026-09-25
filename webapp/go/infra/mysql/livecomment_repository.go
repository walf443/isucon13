package mysql

import (
	"context"
	"database/sql"
	"errors"

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

func (r *livecommentRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id model.LivecommentID) (*model.Livecomment, error) {
	var livecommentModel model.LivecommentModel
	err := q.GetContext(ctx, &livecommentModel, "SELECT id, user_id, livestream_id, comment, tip, created_at FROM livecomments WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	livecomments, err := fillLivecomments(ctx, q, []*model.LivecommentModel{&livecommentModel}, r.fallbackIcon)
	if err != nil {
		return nil, err
	}
	return livecomments[0], nil
}

func (r *livecommentRepository) Create(ctx context.Context, q repository.Querier, livecomment *model.LivecommentModel) (model.LivecommentID, error) {
	rs, err := q.ExecContext(ctx, "INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (?, ?, ?, ?, ?)", livecomment.UserID, livecomment.LivestreamID, livecomment.Comment, livecomment.Tip, livecomment.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := rs.LastInsertId()
	if err != nil {
		return 0, err
	}
	return model.LivecommentID(id), nil
}
