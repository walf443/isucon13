package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// LivecommentReportFiller はライブコメントの報告に報告したユーザ・報告されたライブコメントを埋めた
// domain.LivecommentReportDetail を組み立てる。
type LivecommentReportFiller struct {
	userRepo          repository.UserRepository
	livecommentRepo   repository.LivecommentRepository
	userFiller        *UserFiller
	livecommentFiller *LivecommentFiller
}

func NewLivecommentReportFiller(userRepo repository.UserRepository, livecommentRepo repository.LivecommentRepository, userFiller *UserFiller, livecommentFiller *LivecommentFiller) *LivecommentReportFiller {
	return &LivecommentReportFiller{userRepo: userRepo, livecommentRepo: livecommentRepo, userFiller: userFiller, livecommentFiller: livecommentFiller}
}

// Fill は reports に報告したユーザ・報告されたライブコメントを埋めた domain.LivecommentReportDetail を、報告の ID ごとに返す。
// 同じユーザ・ライブコメントが複数含まれていても 1 回だけ取得する。並び順は呼び出し側で決める。
//
// ユーザやライブコメントが無いのはデータ不整合なので、呼び出し側はこのエラーを報告不在として扱わないこと。
func (f *LivecommentReportFiller) Fill(ctx context.Context, q repository.Querier, reports []*domain.LivecommentReport) (map[domain.LivecommentReportID]*domain.LivecommentReportDetail, error) {
	reporters := make([]*domain.User, 0, len(reports))
	fetchedReporters := make(map[domain.UserID]bool, len(reports))
	livecomments := make([]*domain.Livecomment, 0, len(reports))
	fetchedLivecomments := make(map[domain.LivecommentID]bool, len(reports))
	for _, report := range reports {
		if !fetchedReporters[report.UserID] {
			reporter, err := f.userRepo.FindByID(ctx, q, report.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get reporter of livecomment report %d: %w", report.ID, asMissingDetail(err))
			}
			reporters = append(reporters, reporter)
			fetchedReporters[report.UserID] = true
		}

		if !fetchedLivecomments[report.LivecommentID] {
			livecomment, err := f.livecommentRepo.FindByID(ctx, q, report.LivecommentID)
			if err != nil {
				return nil, fmt.Errorf("failed to get livecomment of livecomment report %d: %w", report.ID, asMissingDetail(err))
			}
			livecomments = append(livecomments, livecomment)
			fetchedLivecomments[report.LivecommentID] = true
		}
	}

	reporterDetails, err := f.userFiller.Fill(ctx, q, reporters)
	if err != nil {
		return nil, err
	}
	livecommentDetails, err := f.livecommentFiller.Fill(ctx, q, livecomments)
	if err != nil {
		return nil, err
	}

	reportDetails := make(map[domain.LivecommentReportID]*domain.LivecommentReportDetail, len(reports))
	for _, report := range reports {
		reportDetails[report.ID] = &domain.LivecommentReportDetail{
			ID:          report.ID,
			Reporter:    *reporterDetails[report.UserID],
			Livecomment: *livecommentDetails[report.LivecommentID],
			CreatedAt:   report.CreatedAt,
		}
	}
	return reportDetails, nil
}
