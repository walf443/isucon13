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

func (r *reactionRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id model.ReactionID) (*model.Reaction, error) {
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

func (r *reactionRepository) FindAllWithDetailsByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.Reaction, error) {
	var reactionModels []*model.ReactionModel
	if err := q.SelectContext(ctx, &reactionModels, "SELECT id, emoji_name, user_id, livestream_id, created_at FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC", livestreamID); err != nil {
		return nil, err
	}
	return fillReactions(ctx, q, reactionModels, r.fallbackIcon)
}

func (r *reactionRepository) FindAllWithDetailsByLivestreamIDLimited(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID, limit int64) ([]*model.Reaction, error) {
	var reactionModels []*model.ReactionModel
	if err := q.SelectContext(ctx, &reactionModels, "SELECT id, emoji_name, user_id, livestream_id, created_at FROM reactions WHERE livestream_id = ? ORDER BY created_at DESC LIMIT ?", livestreamID, limit); err != nil {
		return nil, err
	}
	return fillReactions(ctx, q, reactionModels, r.fallbackIcon)
}

func (r *reactionRepository) Create(ctx context.Context, q repository.Querier, reaction *model.ReactionModel) (model.ReactionID, error) {
	rs, err := q.ExecContext(ctx, "INSERT INTO reactions (user_id, livestream_id, emoji_name, created_at) VALUES (?, ?, ?, ?)", reaction.UserID, reaction.LivestreamID, reaction.EmojiName, reaction.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := rs.LastInsertId()
	if err != nil {
		return 0, err
	}
	return model.ReactionID(id), nil
}

func (r *reactionRepository) CountByLivestreamOwnerID(ctx context.Context, q repository.Querier, userID model.UserID) (int64, error) {
	var reactions int64
	// クエリ文字列は移行前のまま (空白も含めて) にしている
	query := `
		SELECT COUNT(*) FROM users u
		INNER JOIN livestreams l ON l.user_id = u.id
		INNER JOIN reactions r ON r.livestream_id = l.id
		WHERE u.id = ?`
	if err := q.GetContext(ctx, &reactions, query, userID); err != nil {
		return 0, err
	}
	return reactions, nil
}

func (r *reactionRepository) CountByLivestreamOwnerName(ctx context.Context, q repository.Querier, name string) (int64, error) {
	var reactions int64
	// クエリ文字列は移行前のまま (空白も含めて) にしている
	query := `SELECT COUNT(*) FROM users u 
    INNER JOIN livestreams l ON l.user_id = u.id 
    INNER JOIN reactions r ON r.livestream_id = l.id
    WHERE u.name = ?
	`
	if err := q.GetContext(ctx, &reactions, query, name); err != nil {
		return 0, err
	}
	return reactions, nil
}

func (r *reactionRepository) FindFavoriteEmojiByLivestreamOwnerName(ctx context.Context, q repository.Querier, name string) (string, error) {
	var favoriteEmoji string
	// クエリ文字列は移行前のまま (空白も含めて) にしている
	query := `
	SELECT r.emoji_name
	FROM users u
	INNER JOIN livestreams l ON l.user_id = u.id
	INNER JOIN reactions r ON r.livestream_id = l.id
	WHERE u.name = ?
	GROUP BY emoji_name
	ORDER BY COUNT(*) DESC, emoji_name DESC
	LIMIT 1
	`
	err := q.GetContext(ctx, &favoriteEmoji, query, name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", repository.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return favoriteEmoji, nil
}
