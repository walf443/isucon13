package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// fakeLivestreamTagRepository はテストで設定した関数に処理を委ねる LivestreamTagRepository。
// 関数を設定していないメソッドを呼ぶと panic する (埋め込んだインターフェースは nil)。
type fakeLivestreamTagRepository struct {
	repository.LivestreamTagRepository

	findAllByLivestreamID  func(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.LivestreamTag, error)
	findAllByLivestreamIDs func(ctx context.Context, q repository.Querier, livestreamIDs []domain.LivestreamID) ([]*domain.LivestreamTag, error)
	findAllByTagIDs        func(ctx context.Context, q repository.Querier, tagIDs []domain.TagID) ([]*domain.LivestreamTag, error)
	create                 func(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID, tagID domain.TagID) error
}

func (r *fakeLivestreamTagRepository) FindAllByLivestreamID(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID) ([]*domain.LivestreamTag, error) {
	return r.findAllByLivestreamID(ctx, q, livestreamID)
}

func (r *fakeLivestreamTagRepository) FindAllByTagIDs(ctx context.Context, q repository.Querier, tagIDs []domain.TagID) ([]*domain.LivestreamTag, error) {
	return r.findAllByTagIDs(ctx, q, tagIDs)
}

func (r *fakeLivestreamTagRepository) Create(ctx context.Context, q repository.Querier, livestreamID domain.LivestreamID, tagID domain.TagID) error {
	return r.create(ctx, q, livestreamID, tagID)
}

func (r *fakeLivestreamTagRepository) FindAllByLivestreamIDs(ctx context.Context, q repository.Querier, livestreamIDs []domain.LivestreamID) ([]*domain.LivestreamTag, error) {
	return r.findAllByLivestreamIDs(ctx, q, livestreamIDs)
}
