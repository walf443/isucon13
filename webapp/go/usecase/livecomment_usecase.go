package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type LivecommentUsecase interface {
	// FindAllByLivestreamID は指定したライブ配信へのライブコメントを作成日時の降順で返す。
	// limit が nil でなければ最大 *limit 件に絞る。
	FindAllByLivestreamID(ctx context.Context, livestreamID model.LivestreamID, limit *int64) ([]*model.Livecomment, error)
	// FindAllReportsByLivestreamID は指定したライブ配信へのライブコメントの報告を返す。
	// userID のユーザが配信者でない場合 ErrNotLivestreamOwner を返す。
	FindAllReportsByLivestreamID(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) ([]*model.LivecommentReport, error)
}

type livecommentUsecase struct {
	txManager       repository.TxManager
	livestreamRepo  repository.LivestreamRepository
	livecommentRepo repository.LivecommentRepository
	reportRepo      repository.LivecommentReportRepository
}

func NewLivecommentUsecase(txManager repository.TxManager, livestreamRepo repository.LivestreamRepository, livecommentRepo repository.LivecommentRepository, reportRepo repository.LivecommentReportRepository) LivecommentUsecase {
	return &livecommentUsecase{txManager: txManager, livestreamRepo: livestreamRepo, livecommentRepo: livecommentRepo, reportRepo: reportRepo}
}

func (u *livecommentUsecase) FindAllByLivestreamID(ctx context.Context, livestreamID model.LivestreamID, limit *int64) ([]*model.Livecomment, error) {
	var livecomments []*model.Livecomment
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		if limit == nil {
			livecomments, err = u.livecommentRepo.FindAllWithDetailsByLivestreamID(ctx, q, livestreamID)
		} else {
			livecomments, err = u.livecommentRepo.FindAllWithDetailsByLivestreamIDLimited(ctx, q, livestreamID, *limit)
		}
		if err != nil {
			return fmt.Errorf("failed to get livecomments: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return livecomments, nil
}

func (u *livecommentUsecase) FindAllReportsByLivestreamID(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) ([]*model.LivecommentReport, error) {
	var reports []*model.LivecommentReport
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		// ライブ配信が存在しない場合も (404 ではなく) エラーのまま返す (移行前と同じ)
		livestream, err := u.livestreamRepo.FindByID(ctx, q, livestreamID)
		if err != nil {
			return fmt.Errorf("failed to get livestream: %w", err)
		}
		if !livestream.IsOwnedBy(userID) {
			return ErrNotLivestreamOwner
		}

		reports, err = u.reportRepo.FindAllWithDetailsByLivestreamID(ctx, q, livestreamID)
		if err != nil {
			return fmt.Errorf("failed to get livecomment reports: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return reports, nil
}
