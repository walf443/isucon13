package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

// fakeReservationSlotRepository はテストで設定した関数に処理を委ねる ReservationSlotRepository。
// 関数を設定していないメソッドを呼ぶと panic する (埋め込んだインターフェースは nil)。
type fakeReservationSlotRepository struct {
	repository.ReservationSlotRepository

	findAllByRangeForUpdate   func(ctx context.Context, q repository.Querier, period model.ReservationPeriod) ([]*model.ReservationSlotModel, error)
	findSlotByStartAtAndEndAt func(ctx context.Context, q repository.Querier, startAt int64, endAt int64) (int64, error)
	decrementSlotsByRange     func(ctx context.Context, q repository.Querier, period model.ReservationPeriod) error
}

func (r *fakeReservationSlotRepository) FindAllByRangeForUpdate(ctx context.Context, q repository.Querier, period model.ReservationPeriod) ([]*model.ReservationSlotModel, error) {
	return r.findAllByRangeForUpdate(ctx, q, period)
}

func (r *fakeReservationSlotRepository) FindSlotByStartAtAndEndAt(ctx context.Context, q repository.Querier, startAt int64, endAt int64) (int64, error) {
	return r.findSlotByStartAtAndEndAt(ctx, q, startAt, endAt)
}

func (r *fakeReservationSlotRepository) DecrementSlotsByRange(ctx context.Context, q repository.Querier, period model.ReservationPeriod) error {
	return r.decrementSlotsByRange(ctx, q, period)
}
