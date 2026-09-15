use crate::models::id::Id;
use fake::Dummy;

#[derive(Debug, sqlx::FromRow, Dummy)]
pub struct ReservationSlot {
    #[allow(unused)]
    pub id: Id<Self, i64>,
    pub slot: i64,
    pub start_at: i64,
    pub end_at: i64,
}

pub type ReservationSlotId = Id<ReservationSlot, i64>;
