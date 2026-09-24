package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type UserUsecase interface {
	// FindByID はユーザが存在しない場合 ErrUserNotFound を返す。
	FindByID(ctx context.Context, id int64) (*model.User, error)
	// FindByName はユーザが存在しない場合 ErrUserNotFound を返す。
	FindByName(ctx context.Context, name string) (*model.User, error)
}

type userUsecase struct {
	txManager repository.TxManager
	userRepo  repository.UserRepository
}

func NewUserUsecase(txManager repository.TxManager, userRepo repository.UserRepository) UserUsecase {
	return &userUsecase{txManager: txManager, userRepo: userRepo}
}

func (u *userUsecase) FindByID(ctx context.Context, id int64) (*model.User, error) {
	return u.findUser(ctx, func(q repository.Querier) (*model.User, error) {
		return u.userRepo.FindWithDetailsByID(ctx, q, id)
	})
}

func (u *userUsecase) FindByName(ctx context.Context, name string) (*model.User, error) {
	return u.findUser(ctx, func(q repository.Querier) (*model.User, error) {
		return u.userRepo.FindWithDetailsByName(ctx, q, name)
	})
}

func (u *userUsecase) findUser(ctx context.Context, find func(q repository.Querier) (*model.User, error)) (*model.User, error) {
	var user *model.User
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		user, err = find(q)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}
