qbey::qbey_schema!(
    ReservationSlotTable,
    "reservation_slots",
    [id, slot, start_at, end_at,]
);

pub const TABLE_RESERVATION_SLOTS: ReservationSlotTable = ReservationSlotTable::new();
