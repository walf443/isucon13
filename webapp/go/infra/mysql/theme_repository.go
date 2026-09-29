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

type themeRepository struct{}

func NewThemeRepository() repository.ThemeRepository {
	return &themeRepository{}
}

func (r *themeRepository) FindByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) (*domain.Theme, error) {
	var theme domain.Theme
	err := q.GetContext(ctx, &theme, "SELECT id, user_id, dark_mode FROM themes WHERE user_id = ?", userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &theme, nil
}

func (r *themeRepository) Create(ctx context.Context, q repository.Querier, theme *domain.Theme) error {
	_, err := q.ExecContext(ctx, "INSERT INTO themes (user_id, dark_mode) VALUES(?, ?)", theme.UserID, theme.DarkMode)
	return err
}

func (r *themeRepository) FindAllByUserIDs(ctx context.Context, q repository.Querier, userIDs []domain.UserID) ([]*domain.Theme, error) {
	// IN () は作れないので、空の場合はクエリを発行しない
	if len(userIDs) == 0 {
		return []*domain.Theme{}, nil
	}
	query, params, err := sqlx.In("SELECT id, user_id, dark_mode FROM themes WHERE user_id IN (?)", userIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to construct IN query: %w", err)
	}
	var themes []*domain.Theme
	if err := q.SelectContext(ctx, &themes, query, params...); err != nil {
		return nil, err
	}
	return themes, nil
}
