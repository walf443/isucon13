package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type LivestreamUsecase interface {
	// FindByID はライブ配信が存在しない場合 ErrLivestreamNotFound を返す。
	FindByID(ctx context.Context, id model.LivestreamID) (*model.Livestream, error)
	// FindAllByUserID は指定したユーザが配信者のライブ配信を返す。
	FindAllByUserID(ctx context.Context, userID model.UserID) ([]*model.Livestream, error)
	// FindAllByUsername は指定したユーザが配信者のライブ配信を返す。
	// ユーザが存在しない場合 ErrUserNotFound を返す。
	FindAllByUsername(ctx context.Context, username string) ([]*model.Livestream, error)
	// FindAllByTagName は指定した名前のタグが付いたライブ配信を ID の降順で返す。
	// タグが存在しない場合は空のスライスを返す。
	FindAllByTagName(ctx context.Context, tagName string) ([]*model.Livestream, error)
	// FindAll はライブ配信を ID の降順で返す。limit が nil でなければ最大 *limit 件に絞る。
	FindAll(ctx context.Context, limit *int64) ([]*model.Livestream, error)
	// Reserve はライブ配信を予約し、配信者・タグを含めて返す。
	// 予約区間が予約可能期間に掛かっていない場合 ErrBadReservationTimeRange、
	// 予約区間に空きの無い予約枠がある場合 *ReservationSlotUnavailableError を返す。
	Reserve(ctx context.Context, userID model.UserID, input ReserveLivestreamInput) (*model.Livestream, error)
}

// ReserveLivestreamInput は予約するライブ配信の内容。
type ReserveLivestreamInput struct {
	TagIDs       []model.TagID
	Title        string
	Description  string
	PlaylistUrl  string
	ThumbnailUrl string
	StartAt      int64
	EndAt        int64
}

// 予約可能期間 (2023/11/25 10:00 JST からの1年間)
var (
	reservationTermStartAt = time.Date(2023, 11, 25, 1, 0, 0, 0, time.UTC)
	reservationTermEndAt   = time.Date(2024, 11, 25, 1, 0, 0, 0, time.UTC)
)

type livestreamUsecase struct {
	txManager           repository.TxManager
	userRepo            repository.UserRepository
	tagRepo             repository.TagRepository
	livestreamRepo      repository.LivestreamRepository
	reservationSlotRepo repository.ReservationSlotRepository
	logger              Logger
}

func NewLivestreamUsecase(txManager repository.TxManager, userRepo repository.UserRepository, tagRepo repository.TagRepository, livestreamRepo repository.LivestreamRepository, reservationSlotRepo repository.ReservationSlotRepository, logger Logger) LivestreamUsecase {
	return &livestreamUsecase{
		txManager:           txManager,
		userRepo:            userRepo,
		tagRepo:             tagRepo,
		livestreamRepo:      livestreamRepo,
		reservationSlotRepo: reservationSlotRepo,
		logger:              logger,
	}
}

func (u *livestreamUsecase) FindByID(ctx context.Context, id model.LivestreamID) (*model.Livestream, error) {
	var livestream *model.Livestream
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

func (u *livestreamUsecase) FindAllByUserID(ctx context.Context, userID model.UserID) ([]*model.Livestream, error) {
	var livestreams []*model.Livestream
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

func (u *livestreamUsecase) FindAllByUsername(ctx context.Context, username string) ([]*model.Livestream, error) {
	var livestreams []*model.Livestream
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

func (u *livestreamUsecase) FindAllByTagName(ctx context.Context, tagName string) ([]*model.Livestream, error) {
	var livestreams []*model.Livestream
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		tagIDs, err := u.tagRepo.FindIDsByName(ctx, q, tagName)
		if err != nil {
			return fmt.Errorf("failed to get tags: %w", err)
		}
		// 該当するタグが無ければ IN () が作れないので、検索せずに空を返す
		if len(tagIDs) == 0 {
			livestreams = []*model.Livestream{}
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

func (u *livestreamUsecase) FindAll(ctx context.Context, limit *int64) ([]*model.Livestream, error) {
	var livestreams []*model.Livestream
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

func (u *livestreamUsecase) Reserve(ctx context.Context, userID model.UserID, input ReserveLivestreamInput) (*model.Livestream, error) {
	var livestream *model.Livestream
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		// 移行前はトランザクション開始後に期間をチェックしていたので、同じくトランザクション内で行う
		// 2023/11/25 10:00からの１年間の期間内であるかチェック
		var (
			reserveStartAt = time.Unix(input.StartAt, 0)
			reserveEndAt   = time.Unix(input.EndAt, 0)
		)
		if (reserveStartAt.Equal(reservationTermEndAt) || reserveStartAt.After(reservationTermEndAt)) || (reserveEndAt.Equal(reservationTermStartAt) || reserveEndAt.Before(reservationTermStartAt)) {
			return ErrBadReservationTimeRange
		}

		// 予約枠をみて、予約が可能か調べる
		// NOTE: 並列な予約のoverbooking防止にFOR UPDATEが必要
		slots, err := u.reservationSlotRepo.FindAllByRangeForUpdate(ctx, q, input.StartAt, input.EndAt)
		if err != nil {
			u.logger.Warnf("予約枠一覧取得でエラー発生: %+v", err)
			return fmt.Errorf("failed to get reservation_slots: %w", err)
		}
		for _, slot := range slots {
			count, err := u.reservationSlotRepo.FindSlotByStartAtAndEndAt(ctx, q, slot.StartAt, slot.EndAt)
			if err != nil {
				return fmt.Errorf("failed to get reservation_slots: %w", err)
			}
			// 移行前と同じく、ログには FOR UPDATE で取得した時点の残数を出す
			u.logger.Infof("%d ~ %d予約枠の残数 = %d\n", slot.StartAt, slot.EndAt, slot.Slot)
			if count < 1 {
				return &ReservationSlotUnavailableError{StartAt: input.StartAt, EndAt: input.EndAt}
			}
		}

		if err := u.reservationSlotRepo.DecrementSlotsByRange(ctx, q, input.StartAt, input.EndAt); err != nil {
			return fmt.Errorf("failed to update reservation_slot: %w", err)
		}

		livestreamID, err := u.livestreamRepo.Create(ctx, q, &model.LivestreamModel{
			UserID:       userID,
			Title:        input.Title,
			Description:  input.Description,
			PlaylistUrl:  input.PlaylistUrl,
			ThumbnailUrl: input.ThumbnailUrl,
			StartAt:      input.StartAt,
			EndAt:        input.EndAt,
		})
		if err != nil {
			return fmt.Errorf("failed to insert livestream: %w", err)
		}

		// タグ追加
		for _, tagID := range input.TagIDs {
			if err := u.livestreamRepo.AddTag(ctx, q, livestreamID, tagID); err != nil {
				return fmt.Errorf("failed to insert livestream tag: %w", err)
			}
		}

		livestream, err = u.livestreamRepo.FindWithDetailsByID(ctx, q, livestreamID)
		if err != nil {
			return fmt.Errorf("failed to fill livestream: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return livestream, nil
}
