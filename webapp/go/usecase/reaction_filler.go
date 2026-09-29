package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// ReactionFiller はリアクションにユーザ・ライブ配信を埋めた domain.ReactionDetail を組み立てる。
type ReactionFiller struct {
	userRepo         repository.UserRepository
	livestreamRepo   repository.LivestreamRepository
	userFiller       *UserFiller
	livestreamFiller *LivestreamFiller
}

func NewReactionFiller(userRepo repository.UserRepository, livestreamRepo repository.LivestreamRepository, userFiller *UserFiller, livestreamFiller *LivestreamFiller) *ReactionFiller {
	return &ReactionFiller{userRepo: userRepo, livestreamRepo: livestreamRepo, userFiller: userFiller, livestreamFiller: livestreamFiller}
}

// Fill は reactions にユーザ・ライブ配信を埋めた domain.ReactionDetail を、リアクションの ID ごとに返す。
// 同じユーザ・ライブ配信が複数含まれていても 1 回だけ取得する。並び順は呼び出し側で決める。
//
// ユーザやライブ配信が無いのはデータ不整合なので、呼び出し側はこのエラーをリアクション不在として扱わないこと。
func (f *ReactionFiller) Fill(ctx context.Context, q repository.Querier, reactions []*domain.Reaction) (map[domain.ReactionID]*domain.ReactionDetail, error) {
	users := make([]*domain.User, 0, len(reactions))
	fetchedUsers := make(map[domain.UserID]bool, len(reactions))
	livestreams := make([]*domain.Livestream, 0, len(reactions))
	fetchedLivestreams := make(map[domain.LivestreamID]bool, len(reactions))
	for _, reaction := range reactions {
		if !fetchedUsers[reaction.UserID] {
			user, err := f.userRepo.FindByID(ctx, q, reaction.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user of reaction %d: %w", reaction.ID, err)
			}
			users = append(users, user)
			fetchedUsers[reaction.UserID] = true
		}

		if !fetchedLivestreams[reaction.LivestreamID] {
			livestream, err := f.livestreamRepo.FindByID(ctx, q, reaction.LivestreamID)
			if err != nil {
				return nil, fmt.Errorf("failed to get livestream of reaction %d: %w", reaction.ID, err)
			}
			livestreams = append(livestreams, livestream)
			fetchedLivestreams[reaction.LivestreamID] = true
		}
	}

	userDetails, err := f.userFiller.Fill(ctx, q, users)
	if err != nil {
		return nil, err
	}
	livestreamDetails, err := f.livestreamFiller.Fill(ctx, q, livestreams)
	if err != nil {
		return nil, err
	}

	reactionDetails := make(map[domain.ReactionID]*domain.ReactionDetail, len(reactions))
	for _, reaction := range reactions {
		reactionDetails[reaction.ID] = &domain.ReactionDetail{
			ID:         reaction.ID,
			EmojiName:  reaction.EmojiName,
			User:       *userDetails[reaction.UserID],
			Livestream: *livestreamDetails[reaction.LivestreamID],
			CreatedAt:  reaction.CreatedAt,
		}
	}
	return reactionDetails, nil
}
