package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// reservationSlotRow は reservation_slots テーブルの行。
type reservationSlotRow struct {
	ID      int64 `gorm:"column:id;primaryKey"`
	Slot    int64 `gorm:"column:slot"`
	StartAt int64 `gorm:"column:start_at"`
	EndAt   int64 `gorm:"column:end_at"`
}

func (reservationSlotRow) TableName() string { return "reservation_slots" }

func (r *reservationSlotRow) toDomain() *domain.ReservationSlot {
	return &domain.ReservationSlot{ID: domain.ReservationSlotID(r.ID), Slot: r.Slot, StartAt: r.StartAt, EndAt: r.EndAt}
}

type reservationSlotRepository struct{}

func NewReservationSlotRepository() repository.ReservationSlotRepository {
	return &reservationSlotRepository{}
}

func (r *reservationSlotRepository) FindAllByRangeForUpdate(ctx context.Context, q repository.Querier, period domain.ReservationPeriod) ([]*domain.ReservationSlot, error) {
	var rows []*reservationSlotRow
	// NOTE: 並列な予約の overbooking 防止に FOR UPDATE が必要
	err := dbOf(ctx, q).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Select("id, slot, start_at, end_at").
		Where("start_at >= ? AND end_at <= ?", period.StartAt, period.EndAt).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return mapRows(rows, (*reservationSlotRow).toDomain), nil
}

func (r *reservationSlotRepository) FindSlotByStartAtAndEndAt(ctx context.Context, q repository.Querier, startAt int64, endAt int64) (int64, error) {
	var row reservationSlotRow
	err := dbOf(ctx, q).Select("slot").Where("start_at = ? AND end_at = ?", startAt, endAt).Take(&row).Error
	// 該当が無い場合は ErrNotFound に変換せず、移行前と同じ sql.ErrNoRows を返す (500 のレスポンスの本文に出るメッセージを保つ)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, sql.ErrNoRows
	}
	if err != nil {
		return 0, err
	}
	return row.Slot, nil
}

func (r *reservationSlotRepository) DecrementSlotsByRange(ctx context.Context, q repository.Querier, period domain.ReservationPeriod) error {
	return dbOf(ctx, q).
		Model(&reservationSlotRow{}).
		Where("start_at >= ? AND end_at <= ?", period.StartAt, period.EndAt).
		Update("slot", gorm.Expr("slot - 1")).Error
}
