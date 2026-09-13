#[cfg(test)]
mod decrement_slot_between;
#[cfg(test)]
mod find_all_between_for_update;
#[cfg(test)]
mod find_slot_between;

use crate::tables::reservation_slot::ReservationSlotRow;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::reservation_slot::{ReservationSlot, ReservationSlotId};
use isupipe_core::repos::reservation_slot_repository::ReservationSlotRepository;
use toasty::stmt::Value;

#[derive(Clone)]
pub struct ReservationSlotRepositoryInfra {}

#[async_trait]
impl ReservationSlotRepository for ReservationSlotRepositoryInfra {
    async fn find_all_between_for_update<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        start_at: i64,
        end_at: i64,
    ) -> isupipe_core::repos::Result<Vec<ReservationSlot>> {
        // FOR UPDATE is not supported by toasty's query builder
        let rows = toasty::sql::query(
            "SELECT id, slot, start_at, end_at FROM reservation_slots WHERE start_at >= ? AND end_at <= ? FOR UPDATE",
        )
        .bind(start_at)
        .bind(end_at)
        .exec(conn)
        .await?;

        let mut slots = Vec::with_capacity(rows.len());
        for row in rows {
            let fields = row.into_record().fields;
            let [
                Value::I64(id),
                Value::I64(slot),
                Value::I64(start_at),
                Value::I64(end_at),
            ] = fields.as_slice()
            else {
                return Err(toasty::Error::invalid_result(format!(
                    "unexpected reservation_slots row: {fields:?}"
                ))
                .into());
            };
            slots.push(ReservationSlot {
                id: ReservationSlotId::new(*id),
                slot: *slot,
                start_at: *start_at,
                end_at: *end_at,
            });
        }

        Ok(slots)
    }

    async fn find_slot_between<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        start_at: i64,
        end_at: i64,
    ) -> isupipe_core::repos::Result<i64> {
        let row = ReservationSlotRow::filter(ReservationSlotRow::fields().start_at().eq(start_at))
            .filter(ReservationSlotRow::fields().end_at().eq(end_at))
            .one()
            .exec(conn)
            .await?;

        Ok(row.slot)
    }

    async fn decrement_slot_between<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        start_at: i64,
        end_at: i64,
    ) -> isupipe_core::repos::Result<()> {
        ReservationSlotRow::filter(ReservationSlotRow::fields().start_at().ge(start_at))
            .filter(ReservationSlotRow::fields().end_at().le(end_at))
            .update()
            .slot(toasty::stmt::decrement())
            .exec(conn)
            .await?;

        Ok(())
    }
}
