package mysql

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
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
