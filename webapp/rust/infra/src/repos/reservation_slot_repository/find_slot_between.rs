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
    let start_at: i64 = Faker.fake();
    let end_at: i64 = Faker.fake();

    let result = repo.find_slot_between(&mut tx, start_at, end_at).await;
    assert!(result.is_err());
    let err_msg = format!("{}", result.unwrap_err());
    assert!(
        err_msg.contains("no rows returned"),
        "expected RowNotFound, got: {}",
        err_msg
    );
}

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = ReservationSlotRepositoryInfra {};

    let slot: ReservationSlot = Faker.fake();
    {
        let mut ins = qbey(TABLE_RESERVATION_SLOTS.table()).into_insert();
        ins.add_value(&InsertReservationSlotSetup { slot: &slot });
        let (sql, binds) = ins.into_sql();
        bind_qbey_values!(sqlx::query(sqlx::AssertSqlSafe(sql)), binds)
            .execute(&mut *tx)
            .await
            .unwrap();
    }

    let result = repo
        .find_slot_between(&mut tx, slot.start_at, slot.end_at)
        .await
        .unwrap();
    assert_eq!(result, slot.slot)
}
