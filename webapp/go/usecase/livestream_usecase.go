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
	FindByID(ctx context.Context, id domain.LivestreamID) (*domain.Livestream, error)
	// FindAllByUserID は指定したユーザが配信者のライブ配信を返す。
	FindAllByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Livestream, error)
	// FindAllByUsername は指定したユーザが配信者のライブ配信を返す。
	// ユーザが存在しない場合 ErrUserNotFound を返す。
	FindAllByUsername(ctx context.Context, username string) ([]*domain.Livestream, error)
	// FindAllByTagName は指定した名前のタグが付いたライブ配信を ID の降順で返す。
	// タグが存在しない場合は空のスライスを返す。
	FindAllByTagName(ctx context.Context, tagName string) ([]*domain.Livestream, error)
	// FindAll はライブ配信を ID の降順で返す。limit が nil でなければ最大 *limit 件に絞る。
	FindAll(ctx context.Context, limit *domain.Limit) ([]*domain.Livestream, error)
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

// livestreamModelID は orderedBy に渡すための、ライブ配信の ID を取り出す関数。
func livestreamModelID(livestreamModel *domain.LivestreamModel) domain.LivestreamID {
	return livestreamModel.ID
}

// fillLivestreams は livestreamModels に配信者・タグを埋めて、同じ順序で返す。
func (u *livestreamUsecase) fillLivestreams(ctx context.Context, q repository.Querier, livestreamModels []*domain.LivestreamModel) ([]*domain.Livestream, error) {
	filled, err := u.livestreamFiller.Fill(ctx, q, livestreamModels)
	if err != nil {
		return nil, err
	}
	return orderedBy(livestreamModels, livestreamModelID, filled), nil
}

func (u *livestreamUsecase) FindByID(ctx context.Context, id domain.LivestreamID) (*domain.Livestream, error) {
	var livestream *domain.Livestream
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		livestreamModel, err := u.livestreamRepo.FindByID(ctx, q, id)
		// ライブ配信不在 (404) にするのはライブ配信自体が無い場合だけ。配信者やタグの欠損は 500 にする
		if errors.Is(err, repository.ErrNotFound) {
			return ErrLivestreamNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get livestream: %w", err)
		}

		livestreams, err := u.livestreamFiller.Fill(ctx, q, []*domain.LivestreamModel{livestreamModel})
		if err != nil {
			return fmt.Errorf("failed to get livestream: %w", err)
		}
		livestream = livestreams[id]
		return nil
	})
	if err != nil {
		return nil, err
	}
	return livestream, nil
}

func (u *livestreamUsecase) FindAllByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Livestream, error) {
	var livestreams []*domain.Livestream
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		livestreams, err = u.findAllByUserID(ctx, q, userID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return livestreams, nil
}

func (u *livestreamUsecase) FindAllByUsername(ctx context.Context, username string) ([]*domain.Livestream, error) {
	var livestreams []*domain.Livestream
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		userID, err := u.userRepo.FindIDByName(ctx, q, username)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		livestreams, err = u.findAllByUserID(ctx, q, userID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return livestreams, nil
}

func (u *livestreamUsecase) findAllByUserID(ctx context.Context, q repository.Querier, userID domain.UserID) ([]*domain.Livestream, error) {
	livestreamModels, err := u.livestreamRepo.FindAllByUserID(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get livestreams: %w", err)
	}
	livestreams, err := u.fillLivestreams(ctx, q, livestreamModels)
	if err != nil {
		return nil, fmt.Errorf("failed to get livestreams: %w", err)
	}
	return livestreams, nil
}

func (u *livestreamUsecase) FindAllByTagName(ctx context.Context, tagName string) ([]*domain.Livestream, error) {
	var livestreams []*domain.Livestream
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		tagIDs, err := u.tagRepo.FindIDsByName(ctx, q, tagName)
		if err != nil {
			return fmt.Errorf("failed to get tags: %w", err)
		}
		// 該当するタグが無ければ IN () が作れないので、検索せずに空を返す
		if len(tagIDs) == 0 {
			livestreams = []*domain.Livestream{}
			return nil
		}

		// タグの紐付けごとに 1 件 (同じライブ配信に該当するタグが複数あれば重複する)、ライブ配信の ID の降順
		livestreamTags, err := u.livestreamTagRepo.FindAllByTagIDs(ctx, q, tagIDs)
		if err != nil {
			return fmt.Errorf("failed to get livestreams: %w", err)
		}
		livestreamModels := make([]*domain.LivestreamModel, len(livestreamTags))
		fetched := make(map[domain.LivestreamID]*domain.LivestreamModel, len(livestreamTags))
		for i, livestreamTag := range livestreamTags {
			livestreamModel, ok := fetched[livestreamTag.LivestreamID]
			if !ok {
				livestreamModel, err = u.livestreamRepo.FindByID(ctx, q, livestreamTag.LivestreamID)
				if err != nil {
					return fmt.Errorf("failed to get livestreams: failed to get livestream %d: %w", livestreamTag.LivestreamID, err)
				}
				fetched[livestreamTag.LivestreamID] = livestreamModel
			}
			livestreamModels[i] = livestreamModel
		}

		livestreams, err = u.fillLivestreams(ctx, q, livestreamModels)
		if err != nil {
			return fmt.Errorf("failed to get livestreams: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return livestreams, nil
}

func (u *livestreamUsecase) FindAll(ctx context.Context, limit *domain.Limit) ([]*domain.Livestream, error) {
	var livestreams []*domain.Livestream
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var livestreamModels []*domain.LivestreamModel
		var err error
		if limit == nil {
			livestreamModels, err = u.livestreamRepo.FindAllOrderByIDDesc(ctx, q)
		} else {
			livestreamModels, err = u.livestreamRepo.FindAllOrderByIDDescLimited(ctx, q, *limit)
		}
		if err != nil {
			return fmt.Errorf("failed to get livestreams: %w", err)
		}

		livestreams, err = u.fillLivestreams(ctx, q, livestreamModels)
		if err != nil {
			return fmt.Errorf("failed to get livestreams: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return livestreams, nil
}
