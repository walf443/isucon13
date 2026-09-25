package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type fakeTagRepository struct {
	tags []*model.TagModel
	ids  []model.TagID
	err  error

	gotName string
}

func (r *fakeTagRepository) FindIDsByName(ctx context.Context, q repository.Querier, name string) ([]model.TagID, error) {
	r.gotName = name
	return r.ids, r.err
}

func (r *fakeTagRepository) FindAll(ctx context.Context, q repository.Querier) ([]*model.TagModel, error) {
	return r.tags, r.err
}
