#[cfg(test)]
mod decrement_slot_between;
#[cfg(test)]
mod find_all_between_for_update;
#[cfg(test)]
mod find_slot_between;

use crate::sqipe_support::bind_sqipe_values;
use crate::tables::reservation_slot::TABLE_RESERVATION_SLOTS;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::reservation_slot::ReservationSlot;
use isupipe_core::repos::reservation_slot_repository::ReservationSlotRepository;
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
        let t = &TABLE_RESERVATION_SLOTS;
        let mut q = sqipe(t.table_name());
        q.and_where(t.start_at().gte(start_at));
        q.and_where(t.end_at().lte(end_at));
        q.for_update();
        let (sql, binds) = q.to_sql();
        let slots =
            bind_sqipe_values!(sqlx::query_as::<_, ReservationSlot>(&sql), binds)
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
        let t = &TABLE_RESERVATION_SLOTS;
        let mut q = sqipe(t.table_name());
        q.select(&[t.slot()]);
        q.and_where(t.start_at().eq(start_at));
        q.and_where(t.end_at().eq(end_at));
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
        let t = &TABLE_RESERVATION_SLOTS;
        let mut u = sqipe(t.table_name()).into_update();
        u.set_expr(sqipe::SetExpression::new("`slot` = `slot` - 1"));
        u.and_where(t.start_at().gte(start_at));
        u.and_where(t.end_at().lte(end_at));
        let (sql, binds) = u.to_sql();
        bind_sqipe_values!(sqlx::query(&sql), binds)
            .execute(conn)
            .await?;

        Ok(())
    }
}
