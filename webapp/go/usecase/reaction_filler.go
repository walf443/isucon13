package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// ReactionFiller はリアクションにユーザ・ライブ配信を埋めた domain.Reaction を組み立てる。
type ReactionFiller struct {
	userRepo         repository.UserRepository
	livestreamRepo   repository.LivestreamRepository
	userFiller       *UserFiller
	livestreamFiller *LivestreamFiller
}

func NewReactionFiller(userRepo repository.UserRepository, livestreamRepo repository.LivestreamRepository, userFiller *UserFiller, livestreamFiller *LivestreamFiller) *ReactionFiller {
	return &ReactionFiller{userRepo: userRepo, livestreamRepo: livestreamRepo, userFiller: userFiller, livestreamFiller: livestreamFiller}
}

// Fill は reactionModels にユーザ・ライブ配信を埋めた domain.Reaction を、リアクションの ID ごとに返す。
// 同じユーザ・ライブ配信が複数含まれていても 1 回だけ取得する。並び順は呼び出し側で決める。
//
// ユーザやライブ配信が無いのはデータ不整合なので、呼び出し側はこのエラーをリアクション不在として扱わないこと。
func (f *ReactionFiller) Fill(ctx context.Context, q repository.Querier, reactionModels []*domain.ReactionModel) (map[domain.ReactionID]*domain.ReactionDetail, error) {
	userModels := make([]*domain.UserModel, 0, len(reactionModels))
	fetchedUsers := make(map[domain.UserID]bool, len(reactionModels))
	livestreamModels := make([]*domain.LivestreamModel, 0, len(reactionModels))
	fetchedLivestreams := make(map[domain.LivestreamID]bool, len(reactionModels))
	for _, reactionModel := range reactionModels {
		if !fetchedUsers[reactionModel.UserID] {
			userModel, err := f.userRepo.FindByID(ctx, q, reactionModel.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user of reaction %d: %w", reactionModel.ID, err)
			}
			userModels = append(userModels, userModel)
			fetchedUsers[reactionModel.UserID] = true
		}

		if !fetchedLivestreams[reactionModel.LivestreamID] {
			livestreamModel, err := f.livestreamRepo.FindByID(ctx, q, reactionModel.LivestreamID)
			if err != nil {
				return nil, fmt.Errorf("failed to get livestream of reaction %d: %w", reactionModel.ID, err)
			}
			livestreamModels = append(livestreamModels, livestreamModel)
			fetchedLivestreams[reactionModel.LivestreamID] = true
		}
	}

	users, err := f.userFiller.Fill(ctx, q, userModels)
	if err != nil {
		return nil, err
	}
	livestreams, err := f.livestreamFiller.Fill(ctx, q, livestreamModels)
	if err != nil {
		return nil, err
	}

	reactions := make(map[domain.ReactionID]*domain.ReactionDetail, len(reactionModels))
	for _, reactionModel := range reactionModels {
		reactions[reactionModel.ID] = &domain.ReactionDetail{
			ID:         reactionModel.ID,
			EmojiName:  reactionModel.EmojiName,
			User:       *users[reactionModel.UserID],
			Livestream: *livestreams[reactionModel.LivestreamID],
			CreatedAt:  reactionModel.CreatedAt,
		}
	}
	return reactions, nil
}
