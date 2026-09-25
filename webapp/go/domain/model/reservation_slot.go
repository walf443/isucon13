package model

type ReservationSlotID = ID[ReservationSlotModel]

type ReservationSlotModel struct {
	ID      ReservationSlotID `db:"id"`
	Slot    int64             `db:"slot"`
	StartAt int64             `db:"start_at"`
	EndAt   int64             `db:"end_at"`
}
