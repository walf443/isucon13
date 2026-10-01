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
// ユーザ・ライブ配信はそれぞれ 1 回のクエリでまとめて取得する。並び順は呼び出し側で決める。
//
// ユーザやライブ配信が無いのはデータ不整合なので、呼び出し側はこのエラーをライブコメント不在として扱わないこと。
func (f *LivecommentFiller) Fill(ctx context.Context, q repository.Querier, livecomments []*domain.Livecomment) (map[domain.LivecommentID]*domain.LivecommentDetail, error) {
	foundUsers, err := f.userRepo.FindAllByIDs(ctx, q, uniqueKeys(livecomments, func(livecomment *domain.Livecomment) domain.UserID { return livecomment.UserID }))
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	usersByID := indexBy(foundUsers, func(user *domain.User) domain.UserID { return user.ID })

	foundLivestreams, err := f.livestreamRepo.FindAllByIDs(ctx, q, uniqueKeys(livecomments, func(livecomment *domain.Livecomment) domain.LivestreamID { return livecomment.LivestreamID }))
	if err != nil {
		return nil, fmt.Errorf("failed to get livestreams: %w", err)
	}
	livestreamsByID := indexBy(foundLivestreams, func(livestream *domain.Livestream) domain.LivestreamID { return livestream.ID })

	for _, livecomment := range livecomments {
		if _, ok := usersByID[livecomment.UserID]; !ok {
			return nil, fmt.Errorf("failed to get user of livecomment %d: %w", livecomment.ID, missingDetail())
		}
		if _, ok := livestreamsByID[livecomment.LivestreamID]; !ok {
			return nil, fmt.Errorf("failed to get livestream of livecomment %d: %w", livecomment.ID, missingDetail())
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
