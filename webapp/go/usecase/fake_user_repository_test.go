package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type fakeUserRepository struct {
	id          model.UserID
	user        *model.UserModel
	userDetails *model.User
	err         error
	users       []*model.UserModel
	findAllErr  error
	createID    model.UserID
	createErr   error
	gotCreated  *model.UserModel
	// calls は呼ばれたメソッド名を順に記録する
	calls []string

	gotID   model.UserID
	gotName string
}

func (r *fakeUserRepository) Create(ctx context.Context, q repository.Querier, user *model.UserModel) (model.UserID, error) {
	r.calls = append(r.calls, "Create")
	r.gotCreated = user
	return r.createID, r.createErr
}

func (r *fakeUserRepository) FindAll(ctx context.Context, q repository.Querier) ([]*model.UserModel, error) {
	return r.users, r.findAllErr
}

func (r *fakeUserRepository) FindIDByName(ctx context.Context, q repository.Querier, name string) (model.UserID, error) {
	r.gotName = name
	return r.id, r.err
}

func (r *fakeUserRepository) FindByID(ctx context.Context, q repository.Querier, id model.UserID) (*model.UserModel, error) {
	r.gotID = id
	return r.user, r.err
}

func (r *fakeUserRepository) FindByName(ctx context.Context, q repository.Querier, name string) (*model.UserModel, error) {
	r.gotName = name
	return r.user, r.err
}

func (r *fakeUserRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id model.UserID) (*model.User, error) {
	r.calls = append(r.calls, "FindWithDetailsByID")
	r.gotID = id
	return r.userDetails, r.err
}

func (r *fakeUserRepository) FindWithDetailsByName(ctx context.Context, q repository.Querier, name string) (*model.User, error) {
	r.gotName = name
	return r.userDetails, r.err
}
