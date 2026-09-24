package mysql

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type livestreamTagModel struct {
	ID           int64 `db:"id"`
	LivestreamID int64 `db:"livestream_id"`
	TagID        int64 `db:"tag_id"`
}

// fillLivestreams は livestreamModels に配信者・タグを埋めた model.Livestream を、同じ順序で返す。
//
// 配信者やタグが無いのはデータ不整合なので、repository.ErrNotFound (ライブ配信不在) には変換しない。
func fillLivestreams(ctx context.Context, q repository.Querier, livestreamModels []*model.LivestreamModel, fallbackIcon []byte) ([]*model.Livestream, error) {
	ownerModels := make([]*model.UserModel, len(livestreamModels))
	for i, livestreamModel := range livestreamModels {
		var ownerModel model.UserModel
		if err := q.GetContext(ctx, &ownerModel, "SELECT id, name, display_name, description, password FROM users WHERE id = ?", livestreamModel.UserID); err != nil {
			return nil, fmt.Errorf("failed to get owner of livestream %d: %w", livestreamModel.ID, err)
		}
		ownerModels[i] = &ownerModel
	}
	owners, err := fillUsers(ctx, q, ownerModels, fallbackIcon)
	if err != nil {
		return nil, err
	}

	livestreams := make([]*model.Livestream, len(livestreamModels))
	for i, livestreamModel := range livestreamModels {
		var livestreamTagModels []*livestreamTagModel
		if err := q.SelectContext(ctx, &livestreamTagModels, "SELECT id, livestream_id, tag_id FROM livestream_tags WHERE livestream_id = ?", livestreamModel.ID); err != nil {
			return nil, fmt.Errorf("failed to get tags of livestream %d: %w", livestreamModel.ID, err)
		}

		tags := make([]model.TagModel, len(livestreamTagModels))
		for j, livestreamTagModel := range livestreamTagModels {
			if err := q.GetContext(ctx, &tags[j], "SELECT id, name FROM tags WHERE id = ?", livestreamTagModel.TagID); err != nil {
				return nil, fmt.Errorf("failed to get tag %d of livestream %d: %w", livestreamTagModel.TagID, livestreamModel.ID, err)
			}
		}

		livestreams[i] = &model.Livestream{
			ID:           livestreamModel.ID,
			Owner:        *owners[i],
			Title:        livestreamModel.Title,
			Description:  livestreamModel.Description,
			PlaylistUrl:  livestreamModel.PlaylistUrl,
			ThumbnailUrl: livestreamModel.ThumbnailUrl,
			Tags:         tags,
			StartAt:      livestreamModel.StartAt,
			EndAt:        livestreamModel.EndAt,
		}
	}
	return livestreams, nil
}
