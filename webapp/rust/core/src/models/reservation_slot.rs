#[derive(Debug, sqlx::FromRow)]
pub struct ReservationSlotModel {
    #[allow(unused)]
    pub id: i64,
    pub slot: i64,
    pub start_at: i64,
    pub end_at: i64,
}
