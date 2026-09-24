package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type ThemeUsecase interface {
	// FindByUsername はユーザが存在しない場合 ErrUserNotFound を返す。
	FindByUsername(ctx context.Context, username string) (*model.ThemeModel, error)
}

type themeUsecase struct {
	txManager repository.TxManager
	userRepo  repository.UserRepository
	themeRepo repository.ThemeRepository
}

func NewThemeUsecase(txManager repository.TxManager, userRepo repository.UserRepository, themeRepo repository.ThemeRepository) ThemeUsecase {
	return &themeUsecase{txManager: txManager, userRepo: userRepo, themeRepo: themeRepo}
}

func (u *themeUsecase) FindByUsername(ctx context.Context, username string) (*model.ThemeModel, error) {
	var theme *model.ThemeModel
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		userID, err := u.userRepo.FindIDByName(ctx, q, username)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		theme, err = u.themeRepo.FindByUserID(ctx, q, userID)
		if err != nil {
			return fmt.Errorf("failed to get user theme: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return theme, nil
}
