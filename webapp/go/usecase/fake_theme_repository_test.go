package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type fakeThemeRepository struct {
	theme      *model.ThemeModel
	err        error
	gotUserID  model.UserID
	createErr  error
	gotCreated *model.ThemeModel
}

func (r *fakeThemeRepository) Create(ctx context.Context, q repository.Querier, theme *model.ThemeModel) error {
	r.gotCreated = theme
	return r.createErr
}

func (r *fakeThemeRepository) FindByUserID(ctx context.Context, q repository.Querier, userID model.UserID) (*model.ThemeModel, error) {
	r.gotUserID = userID
	return r.theme, r.err
}
