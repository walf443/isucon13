package mysql

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

// fillLivecomments は livecommentModels にコメントしたユーザ・ライブ配信を埋めた model.Livecomment を、同じ順序で返す。
//
// ユーザやライブ配信が無いのはデータ不整合なので、repository.ErrNotFound には変換しない。
func fillLivecomments(ctx context.Context, q repository.Querier, livecommentModels []*model.LivecommentModel, defaultIconHash model.IconHash) ([]*model.Livecomment, error) {
	userModels := make([]*model.UserModel, len(livecommentModels))
	livestreamModels := make([]*model.LivestreamModel, len(livecommentModels))
	for i, livecommentModel := range livecommentModels {
		var userModel model.UserModel
		if err := q.GetContext(ctx, &userModel, "SELECT id, name, display_name, description, password FROM users WHERE id = ?", livecommentModel.UserID); err != nil {
			return nil, fmt.Errorf("failed to get user of livecomment %d: %w", livecommentModel.ID, err)
		}
		userModels[i] = &userModel

		var livestreamModel model.LivestreamModel
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

	livecomments := make([]*model.Livecomment, len(livecommentModels))
	for i, livecommentModel := range livecommentModels {
		livecomments[i] = &model.Livecomment{
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

// fillLivecommentReports は reportModels に報告したユーザ・報告されたライブコメントを埋めた model.LivecommentReport を、同じ順序で返す。
//
// ユーザやライブコメントが無いのはデータ不整合なので、repository.ErrNotFound には変換しない。
func fillLivecommentReports(ctx context.Context, q repository.Querier, reportModels []*model.LivecommentReportModel, defaultIconHash model.IconHash) ([]*model.LivecommentReport, error) {
	reporterModels := make([]*model.UserModel, len(reportModels))
	livecommentModels := make([]*model.LivecommentModel, len(reportModels))
	for i, reportModel := range reportModels {
		var reporterModel model.UserModel
		if err := q.GetContext(ctx, &reporterModel, "SELECT id, name, display_name, description, password FROM users WHERE id = ?", reportModel.UserID); err != nil {
			return nil, fmt.Errorf("failed to get reporter of livecomment report %d: %w", reportModel.ID, err)
		}
		reporterModels[i] = &reporterModel

		var livecommentModel model.LivecommentModel
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

	reports := make([]*model.LivecommentReport, len(reportModels))
	for i, reportModel := range reportModels {
		reports[i] = &model.LivecommentReport{
			ID:          reportModel.ID,
			Reporter:    *reporters[i],
			Livecomment: *livecomments[i],
			CreatedAt:   reportModel.CreatedAt,
		}
	}
	return reports, nil
}
