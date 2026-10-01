package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type LivestreamUsecase interface {
	// FindByID はライブ配信が存在しない場合 ErrLivestreamNotFound を返す。
	FindByID(ctx context.Context, id domain.LivestreamID) (*domain.LivestreamDetail, error)
	// FindAllByUserID は指定したユーザが配信者のライブ配信を返す。
	FindAllByUserID(ctx context.Context, userID domain.UserID) ([]*domain.LivestreamDetail, error)
	// FindAllByUsername は指定したユーザが配信者のライブ配信を返す。
	// ユーザが存在しない場合 ErrUserNotFound を返す。
	FindAllByUsername(ctx context.Context, username domain.Username) ([]*domain.LivestreamDetail, error)
	// FindAllByTagName は指定した名前のタグが付いたライブ配信を ID の降順で返す。
	// タグが存在しない場合は空のスライスを返す。
	FindAllByTagName(ctx context.Context, tagName string) ([]*domain.LivestreamDetail, error)
	// FindAll はライブ配信を ID の降順で返す。limit が nil でなければ最大 *limit 件に絞る。
	FindAll(ctx context.Context, limit *domain.Limit) ([]*domain.LivestreamDetail, error)
}

type livestreamUsecase struct {
	txManager         repository.TxManager
	userRepo          repository.UserRepository
	tagRepo           repository.TagRepository
	livestreamRepo    repository.LivestreamRepository
	livestreamTagRepo repository.LivestreamTagRepository
	livestreamFiller  *LivestreamFiller
}

func NewLivestreamUsecase(txManager repository.TxManager, userRepo repository.UserRepository, tagRepo repository.TagRepository, livestreamRepo repository.LivestreamRepository, livestreamTagRepo repository.LivestreamTagRepository, livestreamFiller *LivestreamFiller) LivestreamUsecase {
	return &livestreamUsecase{
		txManager:         txManager,
		userRepo:          userRepo,
		tagRepo:           tagRepo,
		livestreamRepo:    livestreamRepo,
		livestreamTagRepo: livestreamTagRepo,
		livestreamFiller:  livestreamFiller,
	}
}

// fillLivestreams は livestreams に配信者・タグを埋めて、同じ順序で返す。
func (u *livestreamUsecase) fillLivestreams(ctx context.Context, q repository.Querier, livestreams []*domain.Livestream) ([]*domain.LivestreamDetail, error) {
	filled, err := u.livestreamFiller.Fill(ctx, q, livestreams)
	if err != nil {
		return nil, err
	}
	return orderedBy(livestreams, filled), nil
}

func (u *livestreamUsecase) FindByID(ctx context.Context, id domain.LivestreamID) (*domain.LivestreamDetail, error) {
	var livestreamDetail *domain.LivestreamDetail
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		livestream, err := u.livestreamRepo.FindByID(ctx, q, id)
		// ライブ配信不在 (404) にするのはライブ配信自体が無い場合だけ。配信者やタグの欠損は 500 にする
		if errors.Is(err, repository.ErrNotFound) {
			return ErrLivestreamNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get livestream: %w", err)
		}

		livestreamDetails, err := u.livestreamFiller.Fill(ctx, q, []*domain.Livestream{livestream})
		if err != nil {
			return fmt.Errorf("failed to get livestream: %w", err)
		}
		livestreamDetail = livestreamDetails[id]
		return nil
	})
	if err != nil {
		return nil, err
	}
	return livestreamDetail, nil
}

func (u *livestreamUsecase) FindAllByUserID(ctx context.Context, userID domain.UserID) ([]*domain.LivestreamDetail, error) {
	var livestreamDetails []*domain.LivestreamDetail
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		livestreamDetails, err = u.findAllByUserID(ctx, q, userID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return livestreamDetails, nil
}

func (u *livestreamUsecase) FindAllByUsername(ctx context.Context, username domain.Username) ([]*domain.LivestreamDetail, error) {
	var livestreamDetails []*domain.LivestreamDetail
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		userID, err := u.userRepo.FindIDByName(ctx, q, username)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		livestreamDetails, err = u.findAllByUserID(ctx, q, userID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return livestreamDetails, nil
}

func (u *livestreamUsecase) findAllByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) ([]*domain.LivestreamDetail, error) {
	livestreams, err := u.livestreamRepo.FindAllByUserID(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get livestreams: %w", err)
	}
	livestreamDetails, err := u.fillLivestreams(ctx, q, livestreams)
	if err != nil {
		return nil, fmt.Errorf("failed to get livestreams: %w", err)
	}
	return livestreamDetails, nil
}

func (u *livestreamUsecase) FindAllByTagName(ctx context.Context, tagName string) ([]*domain.LivestreamDetail, error) {
	var livestreamDetails []*domain.LivestreamDetail
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		tagIDs, err := u.tagRepo.FindIDsByName(ctx, q, tagName)
		if err != nil {
			return fmt.Errorf("failed to get tags: %w", err)
		}
		// 該当するタグが無ければ IN () が作れないので、検索せずに空を返す
		if len(tagIDs) == 0 {
			livestreamDetails = []*domain.LivestreamDetail{}
			return nil
		}

		// タグの紐付けごとに 1 件 (同じライブ配信に該当するタグが複数あれば重複する)、ライブ配信の ID の降順
		livestreamTags, err := u.livestreamTagRepo.FindAllByTagIDs(ctx, q, tagIDs)
		if err != nil {
			return fmt.Errorf("failed to get livestreams: %w", err)
		}
		found, err := u.livestreamRepo.FindAllByIDs(ctx, q, uniqueKeys(livestreamTags, func(livestreamTag *domain.LivestreamTag) domain.LivestreamID { return livestreamTag.LivestreamID }))
		if err != nil {
			return fmt.Errorf("failed to get livestreams: %w", err)
		}
		livestreamsByID := indexBy(found, func(livestream *domain.Livestream) domain.LivestreamID { return livestream.ID })
		livestreams := make([]*domain.Livestream, len(livestreamTags))
		for i, livestreamTag := range livestreamTags {
			livestream, ok := livestreamsByID[livestreamTag.LivestreamID]
			if !ok {
				return fmt.Errorf("failed to get livestreams: failed to get livestream %d: %w", livestreamTag.LivestreamID, missingDetail())
			}
			livestreams[i] = livestream
		}

		livestreamDetails, err = u.fillLivestreams(ctx, q, livestreams)
		if err != nil {
			return fmt.Errorf("failed to get livestreams: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return livestreamDetails, nil
}

func (u *livestreamUsecase) FindAll(ctx context.Context, limit *domain.Limit) ([]*domain.LivestreamDetail, error) {
	var livestreamDetails []*domain.LivestreamDetail
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var livestreams []*domain.Livestream
		var err error
		if limit == nil {
			livestreams, err = u.livestreamRepo.FindAllOrderByIDDesc(ctx, q)
		} else {
			livestreams, err = u.livestreamRepo.FindAllOrderByIDDescLimited(ctx, q, *limit)
		}
		if err != nil {
			return fmt.Errorf("failed to get livestreams: %w", err)
		}

		livestreamDetails, err = u.fillLivestreams(ctx, q, livestreams)
		if err != nil {
			return fmt.Errorf("failed to get livestreams: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return livestreamDetails, nil
}
