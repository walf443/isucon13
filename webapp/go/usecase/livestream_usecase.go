package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type LivestreamUsecase interface {
	// FindByID はライブ配信が存在しない場合 ErrLivestreamNotFound を返す。
	FindByID(ctx context.Context, id int64) (*model.Livestream, error)
	// FindAllByUserID は指定したユーザが配信者のライブ配信を返す。
	FindAllByUserID(ctx context.Context, userID int64) ([]*model.Livestream, error)
	// FindAllByUsername は指定したユーザが配信者のライブ配信を返す。
	// ユーザが存在しない場合 ErrUserNotFound を返す。
	FindAllByUsername(ctx context.Context, username string) ([]*model.Livestream, error)
}

type livestreamUsecase struct {
	txManager      repository.TxManager
	userRepo       repository.UserRepository
	livestreamRepo repository.LivestreamRepository
}

func NewLivestreamUsecase(txManager repository.TxManager, userRepo repository.UserRepository, livestreamRepo repository.LivestreamRepository) LivestreamUsecase {
	return &livestreamUsecase{txManager: txManager, userRepo: userRepo, livestreamRepo: livestreamRepo}
}

func (u *livestreamUsecase) FindByID(ctx context.Context, id int64) (*model.Livestream, error) {
	var livestream *model.Livestream
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		livestream, err = u.livestreamRepo.FindWithDetailsByID(ctx, q, id)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrLivestreamNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get livestream: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return livestream, nil
}

func (u *livestreamUsecase) FindAllByUserID(ctx context.Context, userID int64) ([]*model.Livestream, error) {
	var livestreams []*model.Livestream
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		livestreams, err = u.livestreamRepo.FindAllWithDetailsByUserID(ctx, q, userID)
		if err != nil {
			return fmt.Errorf("failed to get livestreams: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return livestreams, nil
}

func (u *livestreamUsecase) FindAllByUsername(ctx context.Context, username string) ([]*model.Livestream, error) {
	var livestreams []*model.Livestream
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		userID, err := u.userRepo.FindIDByName(ctx, q, username)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		livestreams, err = u.livestreamRepo.FindAllWithDetailsByUserID(ctx, q, userID)
		if err != nil {
			return fmt.Errorf("failed to get livestreams: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return livestreams, nil
}
