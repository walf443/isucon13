package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

// LivecommentReportUsecase はライブコメントの (スパムとしての) 報告を扱う。
type LivecommentReportUsecase interface {
	// FindAllByLivestreamID は指定したライブ配信へのライブコメントの報告を返す。
	// userID のユーザが配信者でない場合 ErrNotLivestreamOwner を返す。
	FindAllByLivestreamID(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) ([]*model.LivecommentReport, error)
	// Create はライブコメントを報告し、報告したユーザ・報告されたライブコメントを含めて返す。
	// ライブ配信が存在しない場合 ErrLivestreamNotFound、ライブコメントが存在しない場合 ErrLivecommentNotFound を返す。
	Create(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID, livecommentID model.LivecommentID) (*model.LivecommentReport, error)
}

type livecommentReportUsecase struct {
	txManager       repository.TxManager
	livestreamRepo  repository.LivestreamRepository
	livecommentRepo repository.LivecommentRepository
	reportRepo      repository.LivecommentReportRepository
	// now は現在時刻を返す。テストで差し替えられるようにしている。
	now func() time.Time
}

func NewLivecommentReportUsecase(txManager repository.TxManager, livestreamRepo repository.LivestreamRepository, livecommentRepo repository.LivecommentRepository, reportRepo repository.LivecommentReportRepository) LivecommentReportUsecase {
	return &livecommentReportUsecase{
		txManager:       txManager,
		livestreamRepo:  livestreamRepo,
		livecommentRepo: livecommentRepo,
		reportRepo:      reportRepo,
		now:             time.Now,
	}
}

func (u *livecommentReportUsecase) FindAllByLivestreamID(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) ([]*model.LivecommentReport, error) {
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

func (u *livecommentReportUsecase) Create(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID, livecommentID model.LivecommentID) (*model.LivecommentReport, error) {
	var report *model.LivecommentReport
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		// ライブコメントがそのライブ配信へのものかは確認しない (移行前と同じ)
		_, err := u.livestreamRepo.FindByID(ctx, q, livestreamID)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrLivestreamNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get livestream: %w", err)
		}

		_, err = u.livecommentRepo.FindByID(ctx, q, livecommentID)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrLivecommentNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get livecomment: %w", err)
		}

		reportID, err := u.reportRepo.Create(ctx, q, &model.LivecommentReportModel{
			UserID:        userID,
			LivestreamID:  livestreamID,
			LivecommentID: livecommentID,
			CreatedAt:     u.now().Unix(),
		})
		if err != nil {
			return fmt.Errorf("failed to insert livecomment report: %w", err)
		}

		report, err = u.reportRepo.FindWithDetailsByID(ctx, q, reportID)
		if err != nil {
			return fmt.Errorf("failed to fill livecomment report: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return report, nil
}
