package mysql

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"github.com/jmoiron/sqlx"
)

type livestreamTagRepository struct{}

func NewLivestreamTagRepository() repository.LivestreamTagRepository {
	return &livestreamTagRepository{}
}

func (r *livestreamTagRepository) FindAllByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.LivestreamTagModel, error) {
	var livestreamTags []*domain.LivestreamTagModel
	if err := q.SelectContext(ctx, &livestreamTags, "SELECT id, livestream_id, tag_id FROM livestream_tags WHERE livestream_id = ?", livestreamID); err != nil {
		return nil, err
	}
	return livestreamTags, nil
}

func (r *livestreamTagRepository) Create(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID, tagID domain.TagID) error {
	_, err := q.ExecContext(ctx, "INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (?, ?)", livestreamID, tagID)
	return err
}

func (r *livestreamTagRepository) FindAllByTagIDs(ctx context.Context, q repository.Querier, tagIDs []domain.TagID) ([]*domain.LivestreamTagModel, error) {
	query, params, err := sqlx.In("SELECT id, livestream_id, tag_id FROM livestream_tags WHERE tag_id IN (?) ORDER BY livestream_id DESC", tagIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to construct IN query: %w", err)
	}
	var livestreamTags []*domain.LivestreamTagModel
	if err := q.SelectContext(ctx, &livestreamTags, query, params...); err != nil {
		return nil, err
	}
	return livestreamTags, nil
}
