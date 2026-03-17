use crate::qbey_support::bind_qbey_values;
use crate::repos::reservation_slot_repository::ReservationSlotRepositoryInfra;
use crate::tables::reservation_slot::TABLE_RESERVATION_SLOTS;
use crate::test_support::InsertReservationSlotSetup;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::reservation_slot::ReservationSlot;
use isupipe_core::repos::reservation_slot_repository::ReservationSlotRepository;
use qbey::prelude::*;
use qbey_mysql::qbey;

#[tokio::test]
async fn empty_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = ReservationSlotRepositoryInfra {};
    let slot: ReservationSlot = Faker.fake();
    repo.decrement_slot_between(&mut tx, slot.start_at, slot.end_at)
        .await
        .unwrap();
}

#[tokio::test]
async fn not_empty_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    // deadlock対策
    let t = &TABLE_RESERVATION_SLOTS;
    let mut q = qbey(t.table());
    q.for_update();
    q.select(&[t.id()]);
    let (sql, binds) = q.to_sql();
    bind_qbey_values!(sqlx::query(&sql), binds)
        .fetch_all(&mut *tx)
        .await
        .unwrap();

    let repo = ReservationSlotRepositoryInfra {};
    let mut slot1: ReservationSlot = Faker.fake();
    let mut slot2: ReservationSlot = Faker.fake();
    slot1.end_at = slot1.start_at + 100;
    slot2.start_at = slot1.end_at + 100;
    slot2.end_at = slot2.start_at + 100;

    {
        let mut ins = qbey(TABLE_RESERVATION_SLOTS.table()).into_insert();
        ins.add_value(&InsertReservationSlotSetup { slot: &slot1 });
        ins.add_value(&InsertReservationSlotSetup { slot: &slot2 });
        let (sql, binds) = ins.to_sql();
        bind_qbey_values!(sqlx::query(&sql), binds).execute(&mut *tx).await.unwrap();
    }

    repo.decrement_slot_between(&mut tx, slot1.start_at, slot2.end_at)
        .await
        .unwrap();

    let mut q = qbey(t.table());
    q.and_where(t.id().eq(*slot1.id.inner()));
    let (sql, binds) = q.to_sql();
    let got1: ReservationSlot =
        bind_qbey_values!(sqlx::query_as::<_, ReservationSlot>(&sql), binds)
            .fetch_one(&mut *tx)
            .await
            .unwrap();
    assert_eq!(slot1.slot - 1, got1.slot);

    let mut q = qbey(t.table());
    q.and_where(t.id().eq(*slot2.id.inner()));
    let (sql, binds) = q.to_sql();
    let got2: ReservationSlot =
        bind_qbey_values!(sqlx::query_as::<_, ReservationSlot>(&sql), binds)
            .fetch_one(&mut *tx)
            .await
            .unwrap();
    assert_eq!(slot2.slot - 1, got2.slot);
}
