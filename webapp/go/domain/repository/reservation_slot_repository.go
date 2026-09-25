package repository

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type ReservationSlotRepository interface {
	// FindAllByRangeForUpdate は [startAt, endAt] に収まる予約枠を FOR UPDATE でロックして返す。
	FindAllByRangeForUpdate(ctx context.Context, q Querier, startAt int64, endAt int64) ([]*model.ReservationSlotModel, error)
	// FindSlotByStartAtAndEndAt は開始・終了時刻が一致する予約枠の残数を返す。
	// 該当が無い場合は ErrNotFound に変換せず、sql.ErrNoRows のまま返す。
	FindSlotByStartAtAndEndAt(ctx context.Context, q Querier, startAt int64, endAt int64) (int64, error)
	// DecrementSlotsByRange は [startAt, endAt] に収まる予約枠の残数を 1 ずつ減らす。
	DecrementSlotsByRange(ctx context.Context, q Querier, startAt int64, endAt int64) error
}
