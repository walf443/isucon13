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
}

type livestreamUsecase struct {
	txManager      repository.TxManager
	livestreamRepo repository.LivestreamRepository
}

func NewLivestreamUsecase(txManager repository.TxManager, livestreamRepo repository.LivestreamRepository) LivestreamUsecase {
	return &livestreamUsecase{txManager: txManager, livestreamRepo: livestreamRepo}
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
