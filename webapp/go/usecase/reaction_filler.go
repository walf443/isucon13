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
// ユーザ・ライブ配信はそれぞれ 1 回のクエリでまとめて取得する。並び順は呼び出し側で決める。
//
// ユーザやライブ配信が無いのはデータ不整合なので、呼び出し側はこのエラーをリアクション不在として扱わないこと。
func (f *ReactionFiller) Fill(ctx context.Context, q repository.Querier, reactions []*domain.Reaction) (map[domain.ReactionID]*domain.ReactionDetail, error) {
	foundUsers, err := f.userRepo.FindAllByIDs(ctx, q, uniqueKeys(reactions, func(reaction *domain.Reaction) domain.UserID { return reaction.UserID }))
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	usersByID := indexBy(foundUsers, func(user *domain.User) domain.UserID { return user.ID })

	foundLivestreams, err := f.livestreamRepo.FindAllByIDs(ctx, q, uniqueKeys(reactions, func(reaction *domain.Reaction) domain.LivestreamID { return reaction.LivestreamID }))
	if err != nil {
		return nil, fmt.Errorf("failed to get livestreams: %w", err)
	}
	livestreamsByID := indexBy(foundLivestreams, func(livestream *domain.Livestream) domain.LivestreamID { return livestream.ID })

	for _, reaction := range reactions {
		if _, ok := usersByID[reaction.UserID]; !ok {
			return nil, fmt.Errorf("failed to get user of reaction %d: %w", reaction.ID, missingDetail())
		}
		if _, ok := livestreamsByID[reaction.LivestreamID]; !ok {
			return nil, fmt.Errorf("failed to get livestream of reaction %d: %w", reaction.ID, missingDetail())
		}
	}

	userDetails, err := f.userFiller.Fill(ctx, q, foundUsers)
	if err != nil {
		return nil, err
	}
	livestreamDetails, err := f.livestreamFiller.Fill(ctx, q, foundLivestreams)
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
