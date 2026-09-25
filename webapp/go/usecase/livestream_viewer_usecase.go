package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type LivestreamViewerUsecase interface {
	// Enter はユーザのライブ配信の視聴開始を記録する。ライブ配信の存在は確認しない。
	Enter(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) error
	// Exit はユーザのライブ配信の視聴履歴を削除する。
	Exit(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) error
}

type livestreamViewerUsecase struct {
	txManager  repository.TxManager
	viewerRepo repository.LivestreamViewersHistoryRepository
	// now は現在時刻を返す。テストで差し替えられるようにしている。
	now func() time.Time
}

func NewLivestreamViewerUsecase(txManager repository.TxManager, viewerRepo repository.LivestreamViewersHistoryRepository) LivestreamViewerUsecase {
	return &livestreamViewerUsecase{txManager: txManager, viewerRepo: viewerRepo, now: time.Now}
}

func (u *livestreamViewerUsecase) Enter(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) error {
	return u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		if err := u.viewerRepo.Create(ctx, q, &model.LivestreamViewersHistoryModel{
			UserID:       userID,
			LivestreamID: livestreamID,
			CreatedAt:    u.now().Unix(),
		}); err != nil {
			return fmt.Errorf("failed to insert livestream_view_history: %w", err)
		}
		return nil
	})
}

func (u *livestreamViewerUsecase) Exit(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) error {
	return u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		if err := u.viewerRepo.DeleteByUserIDAndLivestreamID(ctx, q, userID, livestreamID); err != nil {
			return fmt.Errorf("failed to delete livestream_view_history: %w", err)
		}
		return nil
	})
}
