package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// LivestreamFiller はライブ配信に配信者・タグを埋めた domain.LivestreamDetail を組み立てる。
// 複数の usecase で使うので、1 つ作って共有する。
type LivestreamFiller struct {
	userRepo          repository.UserRepository
	livestreamTagRepo repository.LivestreamTagRepository
	tagRepo           repository.TagRepository
	userFiller        *UserFiller
}

func NewLivestreamFiller(userRepo repository.UserRepository, livestreamTagRepo repository.LivestreamTagRepository, tagRepo repository.TagRepository, userFiller *UserFiller) *LivestreamFiller {
	return &LivestreamFiller{userRepo: userRepo, livestreamTagRepo: livestreamTagRepo, tagRepo: tagRepo, userFiller: userFiller}
}

// Fill は livestreams に配信者・タグを埋めた domain.LivestreamDetail を、ライブ配信の ID ごとに返す。
// 同じライブ配信・配信者・タグが複数含まれていても 1 回だけ取得する。並び順は呼び出し側で決める。
//
// 配信者やタグが無いのはデータ不整合なので、呼び出し側はこのエラーをライブ配信不在 (404) として扱わないこと。
func (f *LivestreamFiller) Fill(ctx context.Context, q repository.Querier, livestreams []*domain.Livestream) (map[domain.LivestreamID]*domain.LivestreamDetail, error) {
	owners := make([]*domain.User, 0, len(livestreams))
	fetchedOwners := make(map[domain.UserID]bool, len(livestreams))
	for _, livestream := range livestreams {
		if fetchedOwners[livestream.UserID] {
			continue
		}
		owner, err := f.userRepo.FindByID(ctx, q, livestream.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get owner of livestream %d: %w", livestream.ID, asMissingDetail(err))
		}
		owners = append(owners, owner)
		fetchedOwners[livestream.UserID] = true
	}
	ownerDetails, err := f.userFiller.Fill(ctx, q, owners)
	if err != nil {
		return nil, err
	}

	livestreamDetails := make(map[domain.LivestreamID]*domain.LivestreamDetail, len(livestreams))
	tagsByID := map[domain.TagID]*domain.Tag{}
	for _, livestream := range livestreams {
		if _, ok := livestreamDetails[livestream.ID]; ok {
			continue
		}

		livestreamTags, err := f.livestreamTagRepo.FindAllByLivestreamID(ctx, q, livestream.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tags of livestream %d: %w", livestream.ID, err)
		}
		tags := make([]domain.Tag, len(livestreamTags))
		for i, livestreamTag := range livestreamTags {
			tag, ok := tagsByID[livestreamTag.TagID]
			if !ok {
				tag, err = f.tagRepo.FindByID(ctx, q, livestreamTag.TagID)
				if err != nil {
					return nil, fmt.Errorf("failed to get tag %d of livestream %d: %w", livestreamTag.TagID, livestream.ID, asMissingDetail(err))
				}
				tagsByID[livestreamTag.TagID] = tag
			}
			tags[i] = *tag
		}

		livestreamDetails[livestream.ID] = &domain.LivestreamDetail{
			ID:           livestream.ID,
			Owner:        *ownerDetails[livestream.UserID],
			Title:        livestream.Title,
			Description:  livestream.Description,
			PlaylistUrl:  livestream.PlaylistUrl,
			ThumbnailUrl: livestream.ThumbnailUrl,
			Tags:         tags,
			StartAt:      livestream.StartAt,
			EndAt:        livestream.EndAt,
		}
	}
	return livestreamDetails, nil
}
