package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type fakeLivestreamViewersHistoryRepository struct {
	err error
	// countsByLivestreamID は CountByLivestreamID が返す視聴履歴の件数
	countsByLivestreamID map[model.LivestreamID]int64
	countErr             error
	viewersCount         int64
	viewersCountErr      error
	gotViewersLivestream model.LivestreamID

	gotCreated      *model.LivestreamViewersHistoryModel
	gotUserID       model.UserID
	gotLivestreamID model.LivestreamID
}

func (r *fakeLivestreamViewersHistoryRepository) CountViewersByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	r.gotViewersLivestream = livestreamID
	return r.viewersCount, r.viewersCountErr
}

func (r *fakeLivestreamViewersHistoryRepository) CountByLivestreamID(ctx context.Context, q repository.Querier, livestreamID model.LivestreamID) (int64, error) {
	return r.countsByLivestreamID[livestreamID], r.countErr
}

func (r *fakeLivestreamViewersHistoryRepository) Create(ctx context.Context, q repository.Querier, viewer *model.LivestreamViewersHistoryModel) error {
	r.gotCreated = viewer
	return r.err
}

func (r *fakeLivestreamViewersHistoryRepository) DeleteByUserIDAndLivestreamID(ctx context.Context, q repository.Querier, userID model.UserID, livestreamID model.LivestreamID) error {
	r.gotUserID = userID
	r.gotLivestreamID = livestreamID
	return r.err
}
