package mysql

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type ngWordRepository struct{}

func NewNGWordRepository() repository.NGWordRepository {
	return &ngWordRepository{}
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
