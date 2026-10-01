package domain

type ReservationSlotID = ID[ReservationSlot]

type ReservationSlot struct {
	ID      ReservationSlotID `db:"id"`
	Slot    int64             `db:"slot"`
	StartAt int64             `db:"start_at"`
	EndAt   int64             `db:"end_at"`
}
