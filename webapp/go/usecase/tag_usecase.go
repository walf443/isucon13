package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type TagUsecase interface {
	FindAll(ctx context.Context) ([]*domain.TagModel, error)
}

type tagUsecase struct {
	txManager repository.TxManager
	tagRepo   repository.TagRepository
}

func NewTagUsecase(txManager repository.TxManager, tagRepo repository.TagRepository) TagUsecase {
	return &tagUsecase{txManager: txManager, tagRepo: tagRepo}
}

func (u *tagUsecase) FindAll(ctx context.Context) ([]*domain.TagModel, error) {
	var tags []*domain.TagModel
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		tags, err = u.tagRepo.FindAll(ctx, q)
		if err != nil {
			return fmt.Errorf("failed to get tags: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tags, nil
}
