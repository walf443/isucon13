package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// LivestreamReservationUsecase はライブ配信の予約を扱う。
type LivestreamReservationUsecase interface {
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

type livestreamReservationUsecase struct {
	txManager           repository.TxManager
	livestreamRepo      repository.LivestreamRepository
	reservationSlotRepo repository.ReservationSlotRepository
	logger              Logger
}

func NewLivestreamReservationUsecase(txManager repository.TxManager, livestreamRepo repository.LivestreamRepository, reservationSlotRepo repository.ReservationSlotRepository, logger Logger) LivestreamReservationUsecase {
	return &livestreamReservationUsecase{
		txManager:           txManager,
		livestreamRepo:      livestreamRepo,
		reservationSlotRepo: reservationSlotRepo,
		logger:              logger,
	}
}

func (u *livestreamReservationUsecase) Reserve(ctx context.Context, userID model.UserID, input ReserveLivestreamInput) (*model.Livestream, error) {
	var livestream *model.Livestream
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		// 移行前はトランザクション開始後に期間をチェックしていたので、同じくトランザクション内で行う
		period := model.ReservationPeriod{StartAt: input.StartAt, EndAt: input.EndAt}
		if !period.IsReservable() {
			return ErrBadReservationTimeRange
		}

		// 予約枠をみて、予約が可能か調べる
		// NOTE: 並列な予約のoverbooking防止にFOR UPDATEが必要
		slots, err := u.reservationSlotRepo.FindAllByRangeForUpdate(ctx, q, period)
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
				return &ReservationSlotUnavailableError{Period: period}
			}
		}

		if err := u.reservationSlotRepo.DecrementSlotsByRange(ctx, q, period); err != nil {
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
