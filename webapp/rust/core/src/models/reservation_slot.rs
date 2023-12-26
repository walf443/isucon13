use crate::models::id::Id;

#[derive(Debug, sqlx::FromRow)]
pub struct ReservationSlot {
    #[allow(unused)]
    pub id: Id<Self>,
    pub slot: i64,
    pub start_at: i64,
    pub end_at: i64,
}

pub type ReservationSlotId = Id<ReservationSlot>;
