package mysql

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type tagRepository struct{}

func NewTagRepository() repository.TagRepository {
	return &tagRepository{}
}

func (r *tagRepository) FindAll(ctx context.Context, q repository.Querier) ([]*model.TagModel, error) {
	var tags []*model.TagModel
	if err := q.SelectContext(ctx, &tags, "SELECT id, name FROM tags"); err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *tagRepository) FindIDsByName(ctx context.Context, q repository.Querier, name string) ([]model.TagID, error) {
	var ids []model.TagID
	if err := q.SelectContext(ctx, &ids, "SELECT id FROM tags WHERE name = ?", name); err != nil {
		return nil, err
	}
	return ids, nil
}
