package mysql

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// fillLivecomments は livecommentModels にコメントしたユーザ・ライブ配信を埋めた domain.Livecomment を、同じ順序で返す。
//
// ユーザやライブ配信が無いのはデータ不整合なので、repository.ErrNotFound には変換しない。
func fillLivecomments(ctx context.Context, q repository.Querier, livecommentModels []*domain.LivecommentModel, defaultIconHash domain.IconHash) ([]*domain.Livecomment, error) {
	userModels := make([]*domain.UserModel, len(livecommentModels))
	livestreamModels := make([]*domain.LivestreamModel, len(livecommentModels))
	for i, livecommentModel := range livecommentModels {
		var userModel domain.UserModel
		if err := q.GetContext(ctx, &userModel, "SELECT id, name, display_name, description, password FROM users WHERE id = ?", livecommentModel.UserID); err != nil {
			return nil, fmt.Errorf("failed to get user of livecomment %d: %w", livecommentModel.ID, err)
		}
		userModels[i] = &userModel

		var livestreamModel domain.LivestreamModel
		if err := q.GetContext(ctx, &livestreamModel, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE id = ?", livecommentModel.LivestreamID); err != nil {
			return nil, fmt.Errorf("failed to get livestream of livecomment %d: %w", livecommentModel.ID, err)
		}
		livestreamModels[i] = &livestreamModel
	}

	users, err := fillUsers(ctx, q, userModels, defaultIconHash)
	if err != nil {
		return nil, err
	}
	livestreams, err := fillLivestreams(ctx, q, livestreamModels, defaultIconHash)
	if err != nil {
		return nil, err
	}

	livecomments := make([]*domain.Livecomment, len(livecommentModels))
	for i, livecommentModel := range livecommentModels {
		livecomments[i] = &domain.Livecomment{
			ID:         livecommentModel.ID,
			User:       *users[i],
			Livestream: *livestreams[i],
			Comment:    livecommentModel.Comment,
			Tip:        livecommentModel.Tip,
			CreatedAt:  livecommentModel.CreatedAt,
		}
	}
	return livecomments, nil
}

// fillLivecommentReports は reportModels に報告したユーザ・報告されたライブコメントを埋めた domain.LivecommentReport を、同じ順序で返す。
//
// ユーザやライブコメントが無いのはデータ不整合なので、repository.ErrNotFound には変換しない。
func fillLivecommentReports(ctx context.Context, q repository.Querier, reportModels []*domain.LivecommentReportModel, defaultIconHash domain.IconHash) ([]*domain.LivecommentReport, error) {
	reporterModels := make([]*domain.UserModel, len(reportModels))
	livecommentModels := make([]*domain.LivecommentModel, len(reportModels))
	for i, reportModel := range reportModels {
		var reporterModel domain.UserModel
		if err := q.GetContext(ctx, &reporterModel, "SELECT id, name, display_name, description, password FROM users WHERE id = ?", reportModel.UserID); err != nil {
			return nil, fmt.Errorf("failed to get reporter of livecomment report %d: %w", reportModel.ID, err)
		}
		reporterModels[i] = &reporterModel

		var livecommentModel domain.LivecommentModel
		if err := q.GetContext(ctx, &livecommentModel, "SELECT id, user_id, livestream_id, comment, tip, created_at FROM livecomments WHERE id = ?", reportModel.LivecommentID); err != nil {
			return nil, fmt.Errorf("failed to get livecomment of livecomment report %d: %w", reportModel.ID, err)
		}
		livecommentModels[i] = &livecommentModel
	}

	reporters, err := fillUsers(ctx, q, reporterModels, defaultIconHash)
	if err != nil {
		return nil, err
	}
	livecomments, err := fillLivecomments(ctx, q, livecommentModels, defaultIconHash)
	if err != nil {
		return nil, err
	}

	reports := make([]*domain.LivecommentReport, len(reportModels))
	for i, reportModel := range reportModels {
		reports[i] = &domain.LivecommentReport{
			ID:          reportModel.ID,
			Reporter:    *reporters[i],
			Livecomment: *livecomments[i],
			CreatedAt:   reportModel.CreatedAt,
		}
	}
	return reports, nil
}
