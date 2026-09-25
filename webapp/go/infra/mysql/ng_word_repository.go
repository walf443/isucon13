package mysql

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type ngWordRepository struct{}

func NewNGWordRepository() repository.NGWordRepository {
	return &ngWordRepository{}
}

func (r *ngWordRepository) FindAllByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) ([]*model.NGWordModel, error) {
	var ngWords []*model.NGWordModel
	if err := q.SelectContext(ctx, &ngWords, "SELECT id, user_id, livestream_id, word, created_at FROM ng_words WHERE livestream_id = ?", livestreamID); err != nil {
		return nil, err
	}
	return ngWords, nil
}

func (r *ngWordRepository) FindAllByUserIDAndLivestreamID(ctx context.Context, q repository.Querier, userID model.UserID, livestreamID model.LivestreamID) ([]*model.NGWordModel, error) {
	var ngWords []*model.NGWordModel
	if err := q.SelectContext(ctx, &ngWords, "SELECT id, user_id, livestream_id, word, created_at FROM ng_words WHERE user_id = ? AND livestream_id = ? ORDER BY created_at DESC", userID, livestreamID); err != nil {
		return nil, err
	}
	return ngWords, nil
}

func (r *ngWordRepository) Matches(ctx context.Context, q repository.Querier, comment string, word string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM
		(SELECT ? AS text) AS texts
		INNER JOIN
		(SELECT CONCAT('%', ?, '%')	AS pattern) AS patterns
		ON texts.text LIKE patterns.pattern;
		`
	var hitSpam int
	if err := q.GetContext(ctx, &hitSpam, query, comment, word); err != nil {
		return false, err
	}
	return hitSpam >= 1, nil
}

func (r *ngWordRepository) Create(ctx context.Context, q repository.Querier, ngWord *model.NGWordModel) (model.NGWordID, error) {
	rs, err := q.ExecContext(ctx, "INSERT INTO ng_words(user_id, livestream_id, word, created_at) VALUES (?, ?, ?, ?)", ngWord.UserID, ngWord.LivestreamID, ngWord.Word, ngWord.CreatedAt)
	if err != nil {
		return 0, err
	}
	id, err := rs.LastInsertId()
	if err != nil {
		return 0, err
	}
	return model.NGWordID(id), nil
}
