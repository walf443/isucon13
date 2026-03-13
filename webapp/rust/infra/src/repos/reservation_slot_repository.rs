#[cfg(test)]
mod decrement_slot_between;
#[cfg(test)]
mod find_all_between_for_update;
#[cfg(test)]
mod find_slot_between;

use crate::sqipe_support::bind_sqipe_values;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::reservation_slot::{self, ReservationSlot};
use isupipe_core::repos::reservation_slot_repository::ReservationSlotRepository;
use sqipe::col;
use sqipe_mysql::sqipe;

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
        // FOR UPDATE - not supported by squipe
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
        let mut q = sqipe(reservation_slot::TABLE_NAME);
        q.select(&["slot"]);
        q.and_where(col("start_at").eq(start_at));
        q.and_where(col("end_at").eq(end_at));
        let (sql, binds) = q.to_sql();
        let count = bind_sqipe_values!(sqlx::query_scalar::<_, i64>(&sql), binds)
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
        // UPDATE - not supported by squipe
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
