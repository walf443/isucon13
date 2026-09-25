package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type LivestreamViewersHistoryRepository interface {
	// Create はライブ配信の視聴履歴を登録する。
	Create(ctx context.Context, q Querier, viewer *domain.LivestreamViewersHistoryModel) error
	// DeleteByUserIDAndLivestreamID は指定したユーザのライブ配信の視聴履歴を全て削除する。
	DeleteByUserIDAndLivestreamID(ctx context.Context, q Querier, userID domain.UserID, livestreamID domain.LivestreamID) error
	// CountByLivestreamID はライブ配信の視聴履歴の件数を返す。
	CountByLivestreamID(ctx context.Context, q Querier, livestreamID domain.LivestreamID) (int64, error)
	// CountViewersByLivestreamID は CountByLivestreamID と同じくライブ配信の視聴履歴の件数を返すが、
	// ライブ配信統計で使っていた livestreams と JOIN するクエリのままにしている (ライブ配信が存在しない場合は 0)。
	CountViewersByLivestreamID(ctx context.Context, q Querier, livestreamID domain.LivestreamID) (int64, error)
}
