package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type NGWordUsecase interface {
	// FindAllByLivestreamID は userID のユーザがライブ配信に登録した NG ワードを、作成日時の降順で返す。
	FindAllByLivestreamID(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) ([]*model.NGWordModel, error)
}

type ngWordUsecase struct {
	txManager  repository.TxManager
	ngWordRepo repository.NGWordRepository
}

func NewNGWordUsecase(txManager repository.TxManager, ngWordRepo repository.NGWordRepository) NGWordUsecase {
	return &ngWordUsecase{txManager: txManager, ngWordRepo: ngWordRepo}
}

func (u *ngWordUsecase) FindAllByLivestreamID(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) ([]*model.NGWordModel, error) {
	var ngWords []*model.NGWordModel
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		ngWords, err = u.ngWordRepo.FindAllByUserIDAndLivestreamID(ctx, q, userID, livestreamID)
		if err != nil {
			return fmt.Errorf("failed to get NG words: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ngWords, nil
}
