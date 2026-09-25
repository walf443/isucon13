package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type UserUsecase interface {
	// FindByID はユーザが存在しない場合 ErrUserNotFound を返す。
	FindByID(ctx context.Context, id model.UserID) (*model.User, error)
	// FindByName はユーザが存在しない場合 ErrUserNotFound を返す。
	FindByName(ctx context.Context, name string) (*model.User, error)
	// Login はユーザ名とパスワードを検証し、ユーザを返す。
	// ユーザが存在しないかパスワードが違う場合 ErrInvalidCredentials を返す。
	Login(ctx context.Context, username string, password string) (*model.UserModel, error)
}

type userUsecase struct {
	txManager repository.TxManager
	userRepo  repository.UserRepository
}

func NewUserUsecase(txManager repository.TxManager, userRepo repository.UserRepository) UserUsecase {
	return &userUsecase{txManager: txManager, userRepo: userRepo}
}

func (u *userUsecase) FindByID(ctx context.Context, id model.UserID) (*model.User, error) {
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

func (u *userUsecase) Login(ctx context.Context, username string, password string) (*model.UserModel, error) {
	var user *model.UserModel
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		// usernameはUNIQUEなので、whereで一意に特定できる
		user, err = u.userRepo.FindByName(ctx, q, username)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidCredentials
		}
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 移行前と同じく、トランザクションをコミットしてからパスワードを検証する
	matched, err := user.HashedPassword.Matches(password)
	if err != nil {
		return nil, fmt.Errorf("failed to compare hash and password: %w", err)
	}
	if !matched {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}
