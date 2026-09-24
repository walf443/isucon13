package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type fakeTxManager struct{}

func (m *fakeTxManager) RunInTx(ctx context.Context, fn func(q repository.Querier) error) error {
	return fn(nil)
}

type fakeTagRepository struct {
	tags []*model.TagModel
	err  error
}

func (r *fakeTagRepository) FindAll(ctx context.Context, q repository.Querier) ([]*model.TagModel, error) {
	return r.tags, r.err
}

type fakeUserRepository struct {
	id   int64
	user *model.UserModel
	err  error
}

func (r *fakeUserRepository) FindIDByName(ctx context.Context, q repository.Querier, name string) (int64, error) {
	return r.id, r.err
}

func (r *fakeUserRepository) FindByName(ctx context.Context, q repository.Querier, name string) (*model.UserModel, error) {
	return r.user, r.err
}

type fakeThemeRepository struct {
	theme     *model.ThemeModel
	err       error
	gotUserID int64
}

func (r *fakeThemeRepository) FindByUserID(ctx context.Context, q repository.Querier, userID int64) (*model.ThemeModel, error) {
	r.gotUserID = userID
	return r.theme, r.err
}

type fakeIconRepository struct {
	image []byte
	err   error
}

func (r *fakeIconRepository) FindImageByUserID(ctx context.Context, q repository.Querier, userID int64) ([]byte, error) {
	return r.image, r.err
}
