package mysql

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

// fillReactions は reactionModels にユーザ・ライブ配信を埋めた model.Reaction を、同じ順序で返す。
//
// ユーザやライブ配信が無いのはデータ不整合なので、repository.ErrNotFound には変換しない。
func fillReactions(ctx context.Context, q repository.Querier, reactionModels []*model.ReactionModel, defaultIconHash string) ([]*model.Reaction, error) {
	userModels := make([]*model.UserModel, len(reactionModels))
	livestreamModels := make([]*model.LivestreamModel, len(reactionModels))
	for i, reactionModel := range reactionModels {
		var userModel model.UserModel
		if err := q.GetContext(ctx, &userModel, "SELECT id, name, display_name, description, password FROM users WHERE id = ?", reactionModel.UserID); err != nil {
			return nil, fmt.Errorf("failed to get user of reaction %d: %w", reactionModel.ID, err)
		}
		userModels[i] = &userModel

		var livestreamModel model.LivestreamModel
		if err := q.GetContext(ctx, &livestreamModel, "SELECT id, user_id, title, description, playlist_url, thumbnail_url, start_at, end_at FROM livestreams WHERE id = ?", reactionModel.LivestreamID); err != nil {
			return nil, fmt.Errorf("failed to get livestream of reaction %d: %w", reactionModel.ID, err)
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

	reactions := make([]*model.Reaction, len(reactionModels))
	for i, reactionModel := range reactionModels {
		reactions[i] = &model.Reaction{
			ID:         reactionModel.ID,
			EmojiName:  reactionModel.EmojiName,
			User:       *users[i],
			Livestream: *livestreams[i],
			CreatedAt:  reactionModel.CreatedAt,
		}
	}
	return reactions, nil
}
