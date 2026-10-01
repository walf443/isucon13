package mysql

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"github.com/jmoiron/sqlx"
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

func (r *tagRepository) FindIDsByName(ctx context.Context, q repository.Querier, name string) ([]domain.TagID, error) {
	var ids []domain.TagID
	if err := q.SelectContext(ctx, &ids, "SELECT id FROM tags WHERE name = ?", name); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *tagRepository) FindAllByIDs(ctx context.Context, q repository.Querier, ids []domain.TagID) ([]*domain.Tag, error) {
	// IN () は作れないので、空の場合はクエリを発行しない
	if len(ids) == 0 {
		return []*domain.Tag{}, nil
	}
	query, params, err := sqlx.In("SELECT id, name FROM tags WHERE id IN (?)", ids)
	if err != nil {
		return nil, fmt.Errorf("failed to construct IN query: %w", err)
	}
	var tags []*domain.Tag
	if err := q.SelectContext(ctx, &tags, query, params...); err != nil {
		return nil, err
	}
	return tags, nil
}
