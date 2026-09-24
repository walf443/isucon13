package usecase

import (
	"context"
	"crypto/sha256"
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
	txManager    repository.TxManager
	userRepo     repository.UserRepository
	themeRepo    repository.ThemeRepository
	iconRepo     repository.IconRepository
	fallbackIcon []byte
}

// fallbackIcon はアイコン未登録のユーザに使う画像。
func NewUserUsecase(txManager repository.TxManager, userRepo repository.UserRepository, themeRepo repository.ThemeRepository, iconRepo repository.IconRepository, fallbackIcon []byte) UserUsecase {
	return &userUsecase{
		txManager:    txManager,
		userRepo:     userRepo,
		themeRepo:    themeRepo,
		iconRepo:     iconRepo,
		fallbackIcon: fallbackIcon,
	}
}

func (u *userUsecase) FindByID(ctx context.Context, id int64) (*model.User, error) {
	return u.findUser(ctx, func(q repository.Querier) (*model.UserModel, error) {
		return u.userRepo.FindByID(ctx, q, id)
	})
}

func (u *userUsecase) FindByName(ctx context.Context, name string) (*model.User, error) {
	return u.findUser(ctx, func(q repository.Querier) (*model.UserModel, error) {
		return u.userRepo.FindByName(ctx, q, name)
	})
}

// findUser は find で取得したユーザにテーマ・アイコンを埋めて返す。
func (u *userUsecase) findUser(ctx context.Context, find func(q repository.Querier) (*model.UserModel, error)) (*model.User, error) {
	var user *model.User
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		userModel, err := find(q)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		user, err = u.fillUser(ctx, q, userModel)
		if err != nil {
			return fmt.Errorf("failed to fill user: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userUsecase) fillUser(ctx context.Context, q repository.Querier, userModel *model.UserModel) (*model.User, error) {
	theme, err := u.themeRepo.FindByUserID(ctx, q, userModel.ID)
	if err != nil {
		return nil, err
	}

	image, err := u.iconRepo.FindImageByUserID(ctx, q, userModel.ID)
	if errors.Is(err, repository.ErrNotFound) {
		image = u.fallbackIcon
	} else if err != nil {
		return nil, err
	}

	return &model.User{
		ID:          userModel.ID,
		Name:        userModel.Name,
		DisplayName: userModel.DisplayName,
		Description: userModel.Description,
		Theme:       *theme,
		IconHash:    fmt.Sprintf("%x", sha256.Sum256(image)),
	}, nil
}
