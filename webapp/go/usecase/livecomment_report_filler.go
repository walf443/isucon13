package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// LivecommentReportFiller はライブコメントの報告に報告したユーザ・報告されたライブコメントを埋めた
// domain.LivecommentReport を組み立てる。
type LivecommentReportFiller struct {
	userRepo          repository.UserRepository
	livecommentRepo   repository.LivecommentRepository
	userFiller        *UserFiller
	livecommentFiller *LivecommentFiller
}

func NewLivecommentReportFiller(userRepo repository.UserRepository, livecommentRepo repository.LivecommentRepository, userFiller *UserFiller, livecommentFiller *LivecommentFiller) *LivecommentReportFiller {
	return &LivecommentReportFiller{userRepo: userRepo, livecommentRepo: livecommentRepo, userFiller: userFiller, livecommentFiller: livecommentFiller}
}

// Fill は reportModels に報告したユーザ・報告されたライブコメントを埋めた domain.LivecommentReport を、報告の ID ごとに返す。
// 同じユーザ・ライブコメントが複数含まれていても 1 回だけ取得する。並び順は呼び出し側で決める。
//
// ユーザやライブコメントが無いのはデータ不整合なので、呼び出し側はこのエラーを報告不在として扱わないこと。
func (f *LivecommentReportFiller) Fill(ctx context.Context, q repository.Querier, reportModels []*domain.LivecommentReportModel) (map[domain.LivecommentReportID]*domain.LivecommentReportDetail, error) {
	reporterModels := make([]*domain.UserModel, 0, len(reportModels))
	fetchedReporters := make(map[domain.UserID]bool, len(reportModels))
	livecommentModels := make([]*domain.LivecommentModel, 0, len(reportModels))
	fetchedLivecomments := make(map[domain.LivecommentID]bool, len(reportModels))
	for _, reportModel := range reportModels {
		if !fetchedReporters[reportModel.UserID] {
			reporterModel, err := f.userRepo.FindByID(ctx, q, reportModel.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get reporter of livecomment report %d: %w", reportModel.ID, err)
			}
			reporterModels = append(reporterModels, reporterModel)
			fetchedReporters[reportModel.UserID] = true
		}

		if !fetchedLivecomments[reportModel.LivecommentID] {
			livecommentModel, err := f.livecommentRepo.FindByID(ctx, q, reportModel.LivecommentID)
			if err != nil {
				return nil, fmt.Errorf("failed to get livecomment of livecomment report %d: %w", reportModel.ID, err)
			}
			livecommentModels = append(livecommentModels, livecommentModel)
			fetchedLivecomments[reportModel.LivecommentID] = true
		}
	}

	reporters, err := f.userFiller.Fill(ctx, q, reporterModels)
	if err != nil {
		return nil, err
	}
	livecomments, err := f.livecommentFiller.Fill(ctx, q, livecommentModels)
	if err != nil {
		return nil, err
	}

	reports := make(map[domain.LivecommentReportID]*domain.LivecommentReportDetail, len(reportModels))
	for _, reportModel := range reportModels {
		reports[reportModel.ID] = &domain.LivecommentReportDetail{
			ID:          reportModel.ID,
			Reporter:    *reporters[reportModel.UserID],
			Livecomment: *livecomments[reportModel.LivecommentID],
			CreatedAt:   reportModel.CreatedAt,
		}
	}
	return reports, nil
}
