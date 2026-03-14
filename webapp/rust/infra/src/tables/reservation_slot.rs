use sqipe::{col, Col};

pub struct ReservationSlotTable;

pub const TABLE_RESERVATION_SLOTS: ReservationSlotTable = ReservationSlotTable;

impl ReservationSlotTable {
    pub fn table_name(&self) -> &'static str {
        "reservation_slots"
    }

    pub fn id(&self) -> Col {
        col("id")
    }

    pub fn slot(&self) -> Col {
        col("slot")
    }

    pub fn start_at(&self) -> Col {
        col("start_at")
    }

    pub fn end_at(&self) -> Col {
        col("end_at")
    }
}
