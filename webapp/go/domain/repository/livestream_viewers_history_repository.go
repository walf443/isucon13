package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type LivestreamViewersHistoryRepository interface {
	// Create はライブ配信の視聴履歴を登録する。
	Create(ctx context.Context, q Querier, viewer *model.LivestreamViewersHistoryModel) error
	// DeleteByUserIDAndLivestreamID は指定したユーザのライブ配信の視聴履歴を全て削除する。
	DeleteByUserIDAndLivestreamID(ctx context.Context, q Querier, userID model.UserID, livestreamID model.LivestreamID) error
}
