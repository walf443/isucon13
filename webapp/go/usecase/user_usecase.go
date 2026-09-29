package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type UserUsecase interface {
	// FindByID はユーザが存在しない場合 ErrUserNotFound を返す。
	FindByID(ctx context.Context, id domain.UserID) (*domain.UserDetail, error)
	// FindByName はユーザが存在しない場合 ErrUserNotFound を返す。
	FindByName(ctx context.Context, name string) (*domain.UserDetail, error)
	// Login はユーザ名とパスワードを検証し、ユーザを返す。
	// ユーザが存在しないかパスワードが違う場合 ErrInvalidCredentials を返す。
	Login(ctx context.Context, username string, password string) (*domain.User, error)
}

type userUsecase struct {
	txManager  repository.TxManager
	userRepo   repository.UserRepository
	userFiller *UserFiller
}

func NewUserUsecase(txManager repository.TxManager, userRepo repository.UserRepository, userFiller *UserFiller) UserUsecase {
	return &userUsecase{txManager: txManager, userRepo: userRepo, userFiller: userFiller}
}

func (u *userUsecase) FindByID(ctx context.Context, id domain.UserID) (*domain.UserDetail, error) {
	return u.findUser(ctx, func(q repository.Querier) (*domain.User, error) {
		return u.userRepo.FindByID(ctx, q, id)
	})
}

func (u *userUsecase) FindByName(ctx context.Context, name string) (*domain.UserDetail, error) {
	return u.findUser(ctx, func(q repository.Querier) (*domain.User, error) {
		return u.userRepo.FindByName(ctx, q, name)
	})
}

func (u *userUsecase) findUser(ctx context.Context, find func(q repository.Querier) (*domain.User, error)) (*domain.UserDetail, error) {
	var user *domain.UserDetail
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		userModel, err := find(q)
		// ユーザ不在 (404) にするのはユーザ自体が無い場合だけ。テーマ欠損などは 500 にする
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		users, err := u.userFiller.Fill(ctx, q, []*domain.User{userModel})
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		user = users[userModel.ID]
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userUsecase) Login(ctx context.Context, username string, password string) (*domain.User, error) {
	var user *domain.User
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
