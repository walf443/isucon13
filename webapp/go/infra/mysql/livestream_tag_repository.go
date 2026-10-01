package mysql

import (
	"cmp"
	"context"
	"fmt"
	"slices"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"github.com/jmoiron/sqlx"
)

type livestreamTagRepository struct{}

func NewLivestreamTagRepository() repository.LivestreamTagRepository {
	return &livestreamTagRepository{}
}

func (r *livestreamTagRepository) Create(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID, tagID domain.TagID) error {
	_, err := q.ExecContext(ctx, "INSERT INTO livestream_tags (livestream_id, tag_id) VALUES (?, ?)", livestreamID, tagID)
	return err
}

func (r *livestreamTagRepository) FindAllByTagIDs(ctx context.Context, q repository.Querier, tagIDs []domain.TagID) ([]*domain.LivestreamTag, error) {
	query, params, err := sqlx.In("SELECT id, livestream_id, tag_id FROM livestream_tags WHERE tag_id IN (?) ORDER BY livestream_id DESC", tagIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to construct IN query: %w", err)
	}
	var livestreamTags []*domain.LivestreamTag
	if err := q.SelectContext(ctx, &livestreamTags, query, params...); err != nil {
		return nil, err
	}
	return livestreamTags, nil
}

func (r *livestreamTagRepository) FindAllByLivestreamIDs(ctx context.Context, q repository.Querier, livestreamIDs []domain.LivestreamID) ([]*domain.LivestreamTag, error) {
	livestreamTags, err := selectIn[domain.LivestreamTag](ctx, q, "SELECT id, livestream_id, tag_id FROM livestream_tags WHERE livestream_id IN (?) ORDER BY id", livestreamIDs)
	if err != nil {
		return nil, err
	}
	// ID を分割して引いた場合に、結果全体で ID の昇順になるようにする
	slices.SortFunc(livestreamTags, func(a, b *domain.LivestreamTag) int { return cmp.Compare(a.ID, b.ID) })
	return livestreamTags, nil
}
