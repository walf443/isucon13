package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// LivestreamFiller はライブ配信に配信者・タグを埋めた domain.Livestream を組み立てる。
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

// Fill は livestreamModels に配信者・タグを埋めた domain.Livestream を、ライブ配信の ID ごとに返す。
// 同じライブ配信・配信者・タグが複数含まれていても 1 回だけ取得する。並び順は呼び出し側で決める。
//
// 配信者やタグが無いのはデータ不整合なので、呼び出し側はこのエラーをライブ配信不在 (404) として扱わないこと。
func (f *LivestreamFiller) Fill(ctx context.Context, q repository.Querier, livestreamModels []*domain.Livestream) (map[domain.LivestreamID]*domain.LivestreamDetail, error) {
	ownerModels := make([]*domain.User, 0, len(livestreamModels))
	fetchedOwners := make(map[domain.UserID]bool, len(livestreamModels))
	for _, livestreamModel := range livestreamModels {
		if fetchedOwners[livestreamModel.UserID] {
			continue
		}
		ownerModel, err := f.userRepo.FindByID(ctx, q, livestreamModel.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get owner of livestream %d: %w", livestreamModel.ID, err)
		}
		ownerModels = append(ownerModels, ownerModel)
		fetchedOwners[livestreamModel.UserID] = true
	}
	owners, err := f.userFiller.Fill(ctx, q, ownerModels)
	if err != nil {
		return nil, err
	}

	livestreams := make(map[domain.LivestreamID]*domain.LivestreamDetail, len(livestreamModels))
	tagsByID := map[domain.TagID]*domain.Tag{}
	for _, livestreamModel := range livestreamModels {
		if _, ok := livestreams[livestreamModel.ID]; ok {
			continue
		}

		livestreamTags, err := f.livestreamTagRepo.FindAllByLivestreamID(ctx, q, livestreamModel.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tags of livestream %d: %w", livestreamModel.ID, err)
		}
		tags := make([]domain.Tag, len(livestreamTags))
		for i, livestreamTag := range livestreamTags {
			tag, ok := tagsByID[livestreamTag.TagID]
			if !ok {
				tag, err = f.tagRepo.FindByID(ctx, q, livestreamTag.TagID)
				if err != nil {
					return nil, fmt.Errorf("failed to get tag %d of livestream %d: %w", livestreamTag.TagID, livestreamModel.ID, err)
				}
				tagsByID[livestreamTag.TagID] = tag
			}
			tags[i] = *tag
		}

		livestreams[livestreamModel.ID] = &domain.LivestreamDetail{
			ID:           livestreamModel.ID,
			Owner:        *owners[livestreamModel.UserID],
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
