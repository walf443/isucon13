package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type IconUsecase interface {
	// FindImageByUsername はユーザが存在しない場合 ErrUserNotFound、
	// アイコンが未登録の場合 ErrIconNotFound を返す。
	FindImageByUsername(ctx context.Context, username string) ([]byte, error)
	// Update はユーザのアイコンを置き換え、新しいアイコンの ID を返す。
	Update(ctx context.Context, userID model.UserID, image []byte) (model.IconID, error)
}

type iconUsecase struct {
	txManager repository.TxManager
	userRepo  repository.UserRepository
	iconRepo  repository.IconRepository
}

func NewIconUsecase(txManager repository.TxManager, userRepo repository.UserRepository, iconRepo repository.IconRepository) IconUsecase {
	return &iconUsecase{txManager: txManager, userRepo: userRepo, iconRepo: iconRepo}
}

func (u *iconUsecase) FindImageByUsername(ctx context.Context, username string) ([]byte, error) {
	var image []byte
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		userID, err := u.userRepo.FindIDByName(ctx, q, username)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		image, err = u.iconRepo.FindImageByUserID(ctx, q, userID)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrIconNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get user icon: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (u *iconUsecase) Update(ctx context.Context, userID model.UserID, image []byte) (model.IconID, error) {
	var iconID model.IconID
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		if err := u.iconRepo.DeleteByUserID(ctx, q, userID); err != nil {
			return fmt.Errorf("failed to delete old user icon: %w", err)
		}

		var err error
		iconID, err = u.iconRepo.Create(ctx, q, userID, image)
		if err != nil {
			return fmt.Errorf("failed to insert new user icon: %w", err)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return iconID, nil
}
