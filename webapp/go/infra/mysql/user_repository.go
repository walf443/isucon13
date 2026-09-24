package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type userRepository struct {
	// fallbackIcon はアイコン未登録のユーザに使う画像。
	fallbackIcon []byte
}

func NewUserRepository(fallbackIcon []byte) repository.UserRepository {
	return &userRepository{fallbackIcon: fallbackIcon}
}

func (r *userRepository) FindIDByName(ctx context.Context, q repository.Querier, name string) (int64, error) {
	var id int64
	err := q.GetContext(ctx, &id, "SELECT id FROM users WHERE name = ?", name)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, repository.ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *userRepository) FindByName(ctx context.Context, q repository.Querier, name string) (*model.UserModel, error) {
	var user model.UserModel
	err := q.GetContext(ctx, &user, "SELECT id, name, display_name, description, password FROM users WHERE name = ?", name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, q repository.Querier, id int64) (*model.UserModel, error) {
	var user model.UserModel
	err := q.GetContext(ctx, &user, "SELECT id, name, display_name, description, password FROM users WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindWithDetailsByID(ctx context.Context, q repository.Querier, id int64) (*model.User, error) {
	userModel, err := r.FindByID(ctx, q, id)
	if err != nil {
		return nil, err
	}
	return r.fillUser(ctx, q, userModel)
}

func (r *userRepository) FindWithDetailsByName(ctx context.Context, q repository.Querier, name string) (*model.User, error) {
	userModel, err := r.FindByName(ctx, q, name)
	if err != nil {
		return nil, err
	}
	return r.fillUser(ctx, q, userModel)
}

func (r *userRepository) fillUser(ctx context.Context, q repository.Querier, userModel *model.UserModel) (*model.User, error) {
	users, err := fillUsers(ctx, q, []*model.UserModel{userModel}, r.fallbackIcon)
	if err != nil {
		return nil, err
	}
	return users[0], nil
}
