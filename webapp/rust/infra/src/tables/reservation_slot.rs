use isupipe_core::models::reservation_slot::ReservationSlotId;

qbey::qbey_schema!(
    ReservationSlotTable,
    "reservation_slots",
    [id: ReservationSlotId, slot: i64, start_at: i64, end_at: i64], row = ReservationSlotRow);

pub const TABLE_RESERVATION_SLOTS: ReservationSlotTable = ReservationSlotTable::new();
