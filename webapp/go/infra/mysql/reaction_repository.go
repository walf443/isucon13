package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type reactionRepository struct {
	// fallbackIcon はアイコン未登録のユーザに使う画像。
	fallbackIcon []byte
}

func NewReactionRepository(fallbackIcon []byte) repository.ReactionRepository {
	return &reactionRepository{fallbackIcon: fallbackIcon}
}

func (r *reactionRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id int64) (*model.Reaction, error) {
	var reactionModel model.ReactionModel
	err := q.GetContext(ctx, &reactionModel, "SELECT id, emoji_name, user_id, livestream_id, created_at FROM reactions WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	reactions, err := fillReactions(ctx, q, []*model.ReactionModel{&reactionModel}, r.fallbackIcon)
	if err != nil {
		return nil, err
	}
	return reactions[0], nil
}

func (r *reactionRepository) FindAllWithDetailsByLivestreamID(ctx context.Context, q repository.Querier, livestreamID int64) ([]*model.Reaction, error) {
	var reactionModels []*model.ReactionModel
	if err := q.SelectContext(ctx, &reactionModels, "SELECT id, emoji_name, user_id, livestream_id, created_at FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC", livestreamID); err != nil {
		return nil, err
	}
	return fillReactions(ctx, q, reactionModels, r.fallbackIcon)
}

func (r *reactionRepository) FindAllWithDetailsByLivestreamIDLimited(ctx context.Context, q repository.Querier, livestreamID int64, limit int64) ([]*model.Reaction, error) {
	var reactionModels []*model.ReactionModel
	if err := q.SelectContext(ctx, &reactionModels, "SELECT id, emoji_name, user_id, livestream_id, created_at FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC LIMIT ?", livestreamID, limit); err != nil {
		return nil, err
	}
	return fillReactions(ctx, q, reactionModels, r.fallbackIcon)
}

func (r *reactionRepository) Create(ctx context.Context, q repository.Querier, reaction *model.ReactionModel) (int64, error) {
	rs, err := q.ExecContext(ctx, "INSERT INTO reactions (user_id, livestream_id, emoji_name, created_at) VALUES (?, ?, ?, ?)", reaction.UserID, reaction.LivestreamID, reaction.EmojiName, reaction.CreatedAt)
	if err != nil {
		return 0, err
	}
	return rs.LastInsertId()
}
