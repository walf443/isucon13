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
// 配信者・タグの紐付け・タグはそれぞれ 1 回のクエリでまとめて取得する。並び順は呼び出し側で決める。
// タグは紐付けを登録した順に並べる。
//
// 配信者やタグが無いのはデータ不整合なので errMissingDetail を返す (呼び出し側はライブ配信不在 (404) として扱わないこと)。
func (f *LivestreamFiller) Fill(ctx context.Context, q repository.Querier, livestreams []*domain.Livestream) (map[domain.LivestreamID]*domain.LivestreamDetail, error) {
	ownerIDs := uniqueKeys(livestreams, func(livestream *domain.Livestream) domain.UserID { return livestream.UserID })
	found, err := f.userRepo.FindAllByIDs(ctx, q, ownerIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get owners: %w", err)
	}
	ownersByID := indexBy(found, func(user *domain.User) domain.UserID { return user.ID })
	for _, livestream := range livestreams {
		if _, ok := ownersByID[livestream.UserID]; !ok {
			return nil, fmt.Errorf("failed to get owner of livestream %d: %w", livestream.ID, missingDetail())
		}
	}
	owners := make([]*domain.User, len(ownerIDs))
	for i, ownerID := range ownerIDs {
		owners[i] = ownersByID[ownerID]
	}
	ownerDetails, err := f.userFiller.Fill(ctx, q, owners)
	if err != nil {
		return nil, err
	}

	livestreamIDs := uniqueKeys(livestreams, func(livestream *domain.Livestream) domain.LivestreamID { return livestream.ID })
	livestreamTags, err := f.livestreamTagRepo.FindAllByLivestreamIDs(ctx, q, livestreamIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get livestream tags: %w", err)
	}
	tags, err := f.tagRepo.FindAllByIDs(ctx, q, uniqueKeys(livestreamTags, func(livestreamTag *domain.LivestreamTag) domain.TagID { return livestreamTag.TagID }))
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	tagsByID := indexBy(tags, func(tag *domain.Tag) domain.TagID { return tag.ID })
	// 紐付けは ID の昇順で返るので、ライブ配信ごとに分けても登録した順のまま
	livestreamTagsByLivestreamID := map[domain.LivestreamID][]*domain.LivestreamTag{}
	for _, livestreamTag := range livestreamTags {
		livestreamTagsByLivestreamID[livestreamTag.LivestreamID] = append(livestreamTagsByLivestreamID[livestreamTag.LivestreamID], livestreamTag)
	}

	livestreamDetails := make(map[domain.LivestreamID]*domain.LivestreamDetail, len(livestreamIDs))
	for _, livestream := range livestreams {
		if _, ok := livestreamDetails[livestream.ID]; ok {
			continue
		}

		livestreamTagsOfLivestream := livestreamTagsByLivestreamID[livestream.ID]
		tags := make([]domain.Tag, len(livestreamTagsOfLivestream))
		for i, livestreamTag := range livestreamTagsOfLivestream {
			tag, ok := tagsByID[livestreamTag.TagID]
			if !ok {
				return nil, fmt.Errorf("failed to get tag %d of livestream %d: %w", livestreamTag.TagID, livestream.ID, missingDetail())
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
