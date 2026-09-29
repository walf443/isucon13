package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// LivecommentFiller はライブコメントにコメントしたユーザ・ライブ配信を埋めた domain.LivecommentDetail を組み立てる。
type LivecommentFiller struct {
	userRepo         repository.UserRepository
	livestreamRepo   repository.LivestreamRepository
	userFiller       *UserFiller
	livestreamFiller *LivestreamFiller
}

func NewLivecommentFiller(userRepo repository.UserRepository, livestreamRepo repository.LivestreamRepository, userFiller *UserFiller, livestreamFiller *LivestreamFiller) *LivecommentFiller {
	return &LivecommentFiller{userRepo: userRepo, livestreamRepo: livestreamRepo, userFiller: userFiller, livestreamFiller: livestreamFiller}
}

// Fill は livecomments にユーザ・ライブ配信を埋めた domain.LivecommentDetail を、ライブコメントの ID ごとに返す。
// 同じユーザ・ライブ配信が複数含まれていても 1 回だけ取得する。並び順は呼び出し側で決める。
//
// ユーザやライブ配信が無いのはデータ不整合なので、呼び出し側はこのエラーをライブコメント不在として扱わないこと。
func (f *LivecommentFiller) Fill(ctx context.Context, q repository.Querier, livecomments []*domain.Livecomment) (map[domain.LivecommentID]*domain.LivecommentDetail, error) {
	users := make([]*domain.User, 0, len(livecomments))
	fetchedUsers := make(map[domain.UserID]bool, len(livecomments))
	livestreams := make([]*domain.Livestream, 0, len(livecomments))
	fetchedLivestreams := make(map[domain.LivestreamID]bool, len(livecomments))
	for _, livecomment := range livecomments {
		if !fetchedUsers[livecomment.UserID] {
			user, err := f.userRepo.FindByID(ctx, q, livecomment.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get user of livecomment %d: %w", livecomment.ID, err)
			}
			users = append(users, user)
			fetchedUsers[livecomment.UserID] = true
		}

		if !fetchedLivestreams[livecomment.LivestreamID] {
			livestream, err := f.livestreamRepo.FindByID(ctx, q, livecomment.LivestreamID)
			if err != nil {
				return nil, fmt.Errorf("failed to get livestream of livecomment %d: %w", livecomment.ID, err)
			}
			livestreams = append(livestreams, livestream)
			fetchedLivestreams[livecomment.LivestreamID] = true
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

	livecommentDetails := make(map[domain.LivecommentID]*domain.LivecommentDetail, len(livecomments))
	for _, livecomment := range livecomments {
		livecommentDetails[livecomment.ID] = &domain.LivecommentDetail{
			ID:         livecomment.ID,
			User:       *userDetails[livecomment.UserID],
			Livestream: *livestreamDetails[livecomment.LivestreamID],
			Comment:    livecomment.Comment,
			Tip:        livecomment.Tip,
			CreatedAt:  livecomment.CreatedAt,
		}
	}
	return livecommentDetails, nil
}
