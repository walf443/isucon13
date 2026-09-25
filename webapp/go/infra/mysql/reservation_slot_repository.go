package mysql

import (
	"context"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type reservationSlotRepository struct{}

func NewReservationSlotRepository() repository.ReservationSlotRepository {
	return &reservationSlotRepository{}
}

func (r *reservationSlotRepository) FindAllByRangeForUpdate(ctx context.Context, q repository.Querier, period domain.ReservationPeriod) ([]*domain.ReservationSlotModel, error) {
	var slots []*domain.ReservationSlotModel
	if err := q.SelectContext(ctx, &slots, "SELECT id, slot, start_at, end_at FROM reservation_slots WHERE start_at >= ? AND end_at <= ? FOR UPDATE", period.StartAt, period.EndAt); err != nil {
		return nil, err
	}
	return slots, nil
}

func (r *reservationSlotRepository) FindSlotByStartAtAndEndAt(ctx context.Context, q repository.Querier, startAt int64, endAt int64) (int64, error) {
	var count int64
	if err := q.GetContext(ctx, &count, "SELECT slot FROM reservation_slots WHERE start_at = ? AND end_at = ?", startAt, endAt); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *reservationSlotRepository) DecrementSlotsByRange(ctx context.Context, q repository.Querier, period domain.ReservationPeriod) error {
	_, err := q.ExecContext(ctx, "UPDATE reservation_slots SET slot = slot - 1 WHERE start_at >= ? AND end_at <= ?", period.StartAt, period.EndAt)
	return err
}
