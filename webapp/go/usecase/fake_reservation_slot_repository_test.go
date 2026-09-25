package usecase

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type fakeReservationSlotRepository struct {
	slots []*model.ReservationSlotModel
	// counts は FindSlotByStartAtAndEndAt が返す残数 (キーは開始時刻)
	counts       map[int64]int64
	findAllErr   error
	findSlotErr  error
	decrementErr error

	calls      []string
	gotStartAt int64
	gotEndAt   int64
}

func (r *fakeReservationSlotRepository) FindAllByRangeForUpdate(ctx context.Context, q repository.Querier, startAt int64, endAt int64) ([]*model.ReservationSlotModel, error) {
	r.calls = append(r.calls, "FindAllByRangeForUpdate")
	r.gotStartAt = startAt
	r.gotEndAt = endAt
	return r.slots, r.findAllErr
}

func (r *fakeReservationSlotRepository) FindSlotByStartAtAndEndAt(ctx context.Context, q repository.Querier, startAt int64, endAt int64) (int64, error) {
	r.calls = append(r.calls, "FindSlotByStartAtAndEndAt")
	return r.counts[startAt], r.findSlotErr
}

func (r *fakeReservationSlotRepository) DecrementSlotsByRange(ctx context.Context, q repository.Querier, startAt int64, endAt int64) error {
	r.calls = append(r.calls, "DecrementSlotsByRange")
	r.gotStartAt = startAt
	r.gotEndAt = endAt
	return r.decrementErr
}
