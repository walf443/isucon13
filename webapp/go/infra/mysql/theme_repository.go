package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type themeRepository struct{}

func NewThemeRepository() repository.ThemeRepository {
	return &themeRepository{}
}

func (r *themeRepository) FindByUserID(ctx context.Context, q repository.Querier, userID model.UserID) (*model.ThemeModel, error) {
	var theme model.ThemeModel
	err := q.GetContext(ctx, &theme, "SELECT id, user_id, dark_mode FROM themes WHERE user_id = ?", userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &theme, nil
}

func (r *themeRepository) Create(ctx context.Context, q repository.Querier, theme *model.ThemeModel) error {
	_, err := q.ExecContext(ctx, "INSERT INTO themes (user_id, dark_mode) VALUES(?, ?)", theme.UserID, theme.DarkMode)
	return err
}
