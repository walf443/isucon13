use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::reservation_slot::ReservationSlot;
use isupipe_core::repos::reservation_slot_repository::ReservationSlotRepository;

#[derive(Clone)]
pub struct ReservationSlotRepositoryInfra {}

#[async_trait]
impl ReservationSlotRepository for ReservationSlotRepositoryInfra {
    async fn find_all_between_for_update(
        &self,
        conn: &mut DBConn,
        start_at: i64,
        end_at: i64,
    ) -> isupipe_core::repos::Result<Vec<ReservationSlot>> {
        let slots: Vec<ReservationSlot> = sqlx::query_as(
            "SELECT * FROM reservation_slots WHERE start_at >= ? AND end_at <= ? FOR UPDATE",
        )
        .bind(start_at)
        .bind(end_at)
        .fetch_all(conn)
        .await?;

        Ok(slots)
    }

    async fn find_slot_between(
        &self,
        conn: &mut DBConn,
        start_at: i64,
        end_at: i64,
    ) -> isupipe_core::repos::Result<i64> {
        let count: i64 = sqlx::query_scalar(
            "SELECT slot FROM reservation_slots WHERE start_at = ? AND end_at = ?",
        )
        .bind(start_at)
        .bind(end_at)
        .fetch_one(conn)
        .await?;

        Ok(count)
    }

    async fn decrement_slot_between(
        &self,
        conn: &mut DBConn,
        start_at: i64,
        end_at: i64,
    ) -> isupipe_core::repos::Result<()> {
        sqlx::query(
            "UPDATE reservation_slots SET slot = slot - 1 WHERE start_at >= ? AND end_at <= ?",
        )
        .bind(start_at)
        .bind(end_at)
        .execute(conn)
        .await?;

        Ok(())
    }
}
