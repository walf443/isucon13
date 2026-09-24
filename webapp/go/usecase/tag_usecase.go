package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type TagUsecase interface {
	FindAll(ctx context.Context) ([]*model.TagModel, error)
}

type tagUsecase struct {
	db      repository.Querier
	tagRepo repository.TagRepository
}

func NewTagUsecase(db repository.Querier, tagRepo repository.TagRepository) TagUsecase {
	return &tagUsecase{db: db, tagRepo: tagRepo}
}

func (s *tagUsecase) FindAll(ctx context.Context) ([]*model.TagModel, error) {
	return s.tagRepo.FindAll(ctx, s.db)
}
