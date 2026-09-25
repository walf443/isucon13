package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

// fakeLivestreamViewersHistoryRepository はテストで設定した関数に処理を委ねる LivestreamViewersHistoryRepository。
// 関数を設定していないメソッドを呼ぶと panic する (埋め込んだインターフェースは nil)。
type fakeLivestreamViewersHistoryRepository struct {
	repository.LivestreamViewersHistoryRepository

	create                        func(ctx context.Context, q repository.Querier, viewer *model.LivestreamViewersHistoryModel) error
	deleteByUserIDAndLivestreamID func(ctx context.Context, q repository.Querier, userID model.UserID, livestreamID model.LivestreamID) error
	countByLivestreamID           func(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error)
	countViewersByLivestreamID    func(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error)
}

func (r *fakeLivestreamViewersHistoryRepository) Create(ctx context.Context, q repository.Querier, viewer *model.LivestreamViewersHistoryModel) error {
	return r.create(ctx, q, viewer)
}

func (r *fakeLivestreamViewersHistoryRepository) DeleteByUserIDAndLivestreamID(ctx context.Context, q repository.Querier, userID model.UserID, livestreamID model.LivestreamID) error {
	return r.deleteByUserIDAndLivestreamID(ctx, q, userID, livestreamID)
}

func (r *fakeLivestreamViewersHistoryRepository) CountByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	return r.countByLivestreamID(ctx, q, livestreamID)
}

func (r *fakeLivestreamViewersHistoryRepository) CountViewersByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	return r.countViewersByLivestreamID(ctx, q, livestreamID)
}
