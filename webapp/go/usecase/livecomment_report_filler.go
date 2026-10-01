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
// ユーザ・ライブコメントはそれぞれ 1 回のクエリでまとめて取得する。並び順は呼び出し側で決める。
//
// ユーザやライブコメントが無いのはデータ不整合なので、呼び出し側はこのエラーを報告不在として扱わないこと。
func (f *LivecommentReportFiller) Fill(ctx context.Context, q repository.Querier, reports []*domain.LivecommentReport) (map[domain.LivecommentReportID]*domain.LivecommentReportDetail, error) {
	foundReporters, err := f.userRepo.FindAllByIDs(ctx, q, uniqueKeys(reports, func(report *domain.LivecommentReport) domain.UserID { return report.UserID }))
	if err != nil {
		return nil, fmt.Errorf("failed to get reporters: %w", err)
	}
	reportersByID := indexBy(foundReporters, func(user *domain.User) domain.UserID { return user.ID })

	foundLivecomments, err := f.livecommentRepo.FindAllByIDs(ctx, q, uniqueKeys(reports, func(report *domain.LivecommentReport) domain.LivecommentID { return report.LivecommentID }))
	if err != nil {
		return nil, fmt.Errorf("failed to get livecomments: %w", err)
	}
	livecommentsByID := indexBy(foundLivecomments, func(livecomment *domain.Livecomment) domain.LivecommentID { return livecomment.ID })

	for _, report := range reports {
		if _, ok := reportersByID[report.UserID]; !ok {
			return nil, fmt.Errorf("failed to get reporter of livecomment report %d: %w", report.ID, missingDetail())
		}
		if _, ok := livecommentsByID[report.LivecommentID]; !ok {
			return nil, fmt.Errorf("failed to get livecomment of livecomment report %d: %w", report.ID, missingDetail())
		}
	}

	reporterDetails, err := f.userFiller.Fill(ctx, q, foundReporters)
	if err != nil {
		return nil, err
	}
	livecommentDetails, err := f.livecommentFiller.Fill(ctx, q, foundLivecomments)
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
