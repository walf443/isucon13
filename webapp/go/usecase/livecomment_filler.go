package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// LivecommentFiller はライブコメントにコメントしたユーザ・ライブ配信を埋めた domain.Livecomment を組み立てる。
type LivecommentFiller struct {
	userRepo         repository.UserRepository
	livestreamRepo   repository.LivestreamRepository
	userFiller       *UserFiller
	livestreamFiller *LivestreamFiller
}

func NewLivecommentFiller(userRepo repository.UserRepository, livestreamRepo repository.LivestreamRepository, userFiller *UserFiller, livestreamFiller *LivestreamFiller) *LivecommentFiller {
	return &LivecommentFiller{userRepo: userRepo, livestreamRepo: livestreamRepo, userFiller: userFiller, livestreamFiller: livestreamFiller}
}

// Fill は livecommentModels にユーザ・ライブ配信を埋めた domain.Livecomment を、ライブコメントの ID ごとに返す。
// 同じユーザ・ライブ配信が複数含まれていても 1 回だけ取得する。並び順は呼び出し側で決める。
//
// ユーザやライブ配信が無いのはデータ不整合なので、呼び出し側はこのエラーをライブコメント不在として扱わないこと。
func (f *LivecommentFiller) Fill(ctx context.Context, q repository.Querier, livecommentModels []*domain.LivecommentModel) (map[domain.LivecommentID]*domain.Livecomment, error) {
	userModels := make([]*domain.UserModel, 0, len(livecommentModels))
	fetchedUsers := make(map[domain.UserID]bool, len(livecommentModels))
	livestreamModels := make([]*domain.LivestreamModel, 0, len(livecommentModels))
	fetchedLivestreams := make(map[domain.LivestreamID]bool, len(livecommentModels))
	for _, livecommentModel := range livecommentModels {
		if !fetchedUsers[livecommentModel.UserID] {
			userModel, err := f.userRepo.FindByID(ctx, q, livecommentModel.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user of livecomment %d: %w", livecommentModel.ID, err)
			}
			userModels = append(userModels, userModel)
			fetchedUsers[livecommentModel.UserID] = true
		}

		if !fetchedLivestreams[livecommentModel.LivestreamID] {
			livestreamModel, err := f.livestreamRepo.FindByID(ctx, q, livecommentModel.LivestreamID)
			if err != nil {
				return nil, fmt.Errorf("failed to get livestream of livecomment %d: %w", livecommentModel.ID, err)
			}
			livestreamModels = append(livestreamModels, livestreamModel)
			fetchedLivestreams[livecommentModel.LivestreamID] = true
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

	livecomments := make(map[domain.LivecommentID]*domain.Livecomment, len(livecommentModels))
	for _, livecommentModel := range livecommentModels {
		livecomments[livecommentModel.ID] = &domain.Livecomment{
			ID:         livecommentModel.ID,
			User:       *users[livecommentModel.UserID],
			Livestream: *livestreams[livecommentModel.LivestreamID],
			Comment:    livecommentModel.Comment,
			Tip:        livecommentModel.Tip,
			CreatedAt:  livecommentModel.CreatedAt,
		}
	}
	return livecomments, nil
}
