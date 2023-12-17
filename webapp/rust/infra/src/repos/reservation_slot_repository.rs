use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::reservation_slot::ReservationSlotModel;
use isupipe_core::repos::reservation_slot_repository::ReservationSlotRepository;

pub struct ReservationSlotRepositoryInfra {}

#[async_trait]
impl ReservationSlotRepository for ReservationSlotRepositoryInfra {
    async fn find_all_between_for_update(
        &self,
        conn: &mut DBConn,
        start_at: i64,
        end_at: i64,
    ) -> isupipe_core::repos::Result<Vec<ReservationSlotModel>> {
        let slots: Vec<ReservationSlotModel> = sqlx::query_as(
            "SELECT * FROM reservation_slots WHERE start_at >= ? AND end_at <= ? FOR UPDATE",
        )
        .bind(start_at)
        .bind(end_at)
        .fetch_all(conn)
        .await?;

        Ok(slots)
    }
}
