package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type tagRepository struct{}

func NewTagRepository() repository.TagRepository {
	return &tagRepository{}
}

func (r *tagRepository) FindAll(ctx context.Context, q repository.Querier) ([]*domain.Tag, error) {
	var tags []*domain.Tag
	if err := q.SelectContext(ctx, &tags, "SELECT id, name FROM tags"); err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *tagRepository) FindByID(ctx context.Context, q repository.Querier, id domain.TagID) (*domain.Tag, error) {
	var tag domain.Tag
	err := q.GetContext(ctx, &tag, "SELECT id, name FROM tags WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) FindIDsByName(ctx context.Context, q repository.Querier, name string) ([]domain.TagID, error) {
	var ids []domain.TagID
	if err := q.SelectContext(ctx, &ids, "SELECT id FROM tags WHERE name = ?", name); err != nil {
		return nil, err
	}
	return ids, nil
}
