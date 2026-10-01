package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type LivestreamTagRepository interface {
	// FindAllByLivestreamIDs は livestreamIDs のライブ配信に付いたタグの紐付けを、紐付けの ID の昇順で返す。
	// livestreamIDs が空の場合はクエリを発行せずに空のスライスを返す。
	FindAllByLivestreamIDs(ctx context.Context, q Querier, livestreamIDs []domain.LivestreamID) ([]*domain.LivestreamTag, error)
	// FindAllByTagIDs は指定したタグのいずれかの紐付けを、ライブ配信の ID の降順で返す。
	// tagIDs は空であってはならない。
	FindAllByTagIDs(ctx context.Context, q Querier, tagIDs []domain.TagID) ([]*domain.LivestreamTag, error)
	// Create はライブ配信にタグを付ける。
	Create(ctx context.Context, q Querier, livestreamID domain.LivestreamID, tagID domain.TagID) error
}
