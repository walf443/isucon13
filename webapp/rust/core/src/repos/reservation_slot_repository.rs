use crate::db::DBConn;
use crate::models::reservation_slot::ReservationSlotModel;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait ReservationSlotRepository {
    async fn find_all_between_for_update(
        &self,
        conn: &mut DBConn,
        start_at: i64,
        end_at: i64,
    ) -> Result<Vec<ReservationSlotModel>>;

    async fn find_slot_between(&self, conn: &mut DBConn, start_at: i64, end_at: i64)
        -> Result<i64>;
}
