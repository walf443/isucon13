use isupipe_core::models::reservation_slot::{ReservationSlot, ReservationSlotId};

#[derive(Debug, toasty::Model)]
#[table = "reservation_slots"]
pub struct ReservationSlotRow {
    #[key]
    #[auto]
    pub id: ReservationSlotId,
    pub slot: i64,
    pub start_at: i64,
    pub end_at: i64,
}

impl From<ReservationSlotRow> for ReservationSlot {
    fn from(row: ReservationSlotRow) -> Self {
        Self {
            id: row.id,
            slot: row.slot,
            start_at: row.start_at,
            end_at: row.end_at,
        }
    }
}
