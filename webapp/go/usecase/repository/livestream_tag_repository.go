package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type LivestreamTagRepository interface {
	// FindAllByLivestreamID は指定したライブ配信に付いたタグの紐付けを返す (順序は不定)。
	FindAllByLivestreamID(ctx context.Context, q Querier, livestreamID domain.LivestreamID) ([]*domain.LivestreamTag, error)
	// FindAllByTagIDs は指定したタグのいずれかの紐付けを、ライブ配信の ID の降順で返す。
	// tagIDs は空であってはならない。
	FindAllByTagIDs(ctx context.Context, q Querier, tagIDs []domain.TagID) ([]*domain.LivestreamTag, error)
	// Create はライブ配信にタグを付ける。
	Create(ctx context.Context, q Querier, livestreamID domain.LivestreamID, tagID domain.TagID) error
}
