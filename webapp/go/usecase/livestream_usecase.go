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
	txManager      repository.TxManager
	userRepo       repository.UserRepository
	tagRepo        repository.TagRepository
	livestreamRepo repository.LivestreamRepository
}

func NewLivestreamUsecase(txManager repository.TxManager, userRepo repository.UserRepository, tagRepo repository.TagRepository, livestreamRepo repository.LivestreamRepository) LivestreamUsecase {
	return &livestreamUsecase{
		txManager:      txManager,
		userRepo:       userRepo,
		tagRepo:        tagRepo,
		livestreamRepo: livestreamRepo,
	}
}

func (u *livestreamUsecase) FindByID(ctx context.Context, id domain.LivestreamID) (*domain.Livestream, error) {
	var livestream *domain.Livestream
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		livestream, err = u.livestreamRepo.FindWithDetailsByID(ctx, q, id)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrLivestreamNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get livestream: %w", err)
		}
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
		livestreams, err = u.livestreamRepo.FindAllWithDetailsByUserID(ctx, q, userID)
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

		livestreams, err = u.livestreamRepo.FindAllWithDetailsByUserID(ctx, q, userID)
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

		livestreams, err = u.livestreamRepo.FindAllWithDetailsByTagIDs(ctx, q, tagIDs)
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
		var err error
		if limit == nil {
			livestreams, err = u.livestreamRepo.FindAllWithDetails(ctx, q)
		} else {
			livestreams, err = u.livestreamRepo.FindAllWithDetailsLimited(ctx, q, *limit)
		}
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
