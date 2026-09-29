package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"github.com/jmoiron/sqlx"
)

type iconRepository struct{}

func NewIconRepository() repository.IconRepository {
	return &iconRepository{}
}

func (r *iconRepository) FindImageByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) ([]byte, error) {
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

func (r *iconRepository) Create(ctx context.Context, q repository.Querier, userID domain.UserID, image []byte) (domain.IconID, error) {
	rs, err := q.ExecContext(ctx, "INSERT INTO icons (user_id, image) VALUES (?, ?)", userID, image)
	if err != nil {
		return 0, err
	}
	id, err := rs.LastInsertId()
	if err != nil {
		return 0, err
	}
	return domain.IconID(id), nil
}

func (r *iconRepository) DeleteByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) error {
	_, err := q.ExecContext(ctx, "DELETE FROM icons WHERE user_id = ?", userID)
	return err
}

func (r *iconRepository) FindAllByUserIDs(ctx context.Context, q repository.Querier, userIDs []domain.UserID) ([]*domain.Icon, error) {
	// IN () は作れないので、空の場合はクエリを発行しない
	if len(userIDs) == 0 {
		return []*domain.Icon{}, nil
	}
	query, params, err := sqlx.In("SELECT id, user_id, image FROM icons WHERE user_id IN (?)", userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to construct IN query: %w", err)
	}
	var icons []*domain.Icon
	if err := q.SelectContext(ctx, &icons, query, params...); err != nil {
		return nil, err
	}
	return icons, nil
}
