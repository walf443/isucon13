package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type iconRepository struct{}

func NewIconRepository() repository.IconRepository {
	return &iconRepository{}
}

func (r *iconRepository) FindImageByUserID(ctx context.Context, q repository.Querier, userID int64) ([]byte, error) {
	var image []byte
	err := q.GetContext(ctx, &image, "SELECT image FROM icons WHERE user_id = ?", userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (r *iconRepository) Create(ctx context.Context, q repository.Querier, userID int64, image []byte) (int64, error) {
	rs, err := q.ExecContext(ctx, "INSERT INTO icons (user_id, image) VALUES (?, ?)", userID, image)
	if err != nil {
		return 0, err
	}
	return rs.LastInsertId()
}

func (r *iconRepository) DeleteByUserID(ctx context.Context, q repository.Querier, userID int64) error {
	_, err := q.ExecContext(ctx, "DELETE FROM icons WHERE user_id = ?", userID)
	return err
}
