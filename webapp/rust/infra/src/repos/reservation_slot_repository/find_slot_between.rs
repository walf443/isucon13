use crate::repos::reservation_slot_repository::ReservationSlotRepositoryInfra;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::reservation_slot::ReservationSlot;
use isupipe_core::repos::reservation_slot_repository::ReservationSlotRepository;

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
    assert!(err_msg.contains("no rows returned"), "expected RowNotFound, got: {}", err_msg);
}

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = ReservationSlotRepositoryInfra {};

    let slot: ReservationSlot = Faker.fake();
    sqlx::query("INSERT INTO reservation_slots (id, slot, start_at, end_at) VALUES (?, ?, ?, ?)")
        .bind(&slot.id)
        .bind(slot.slot)
        .bind(slot.start_at)
        .bind(slot.end_at)
        .execute(&mut *tx)
        .await
        .unwrap();

    let result = repo
        .find_slot_between(&mut tx, slot.start_at, slot.end_at)
        .await
        .unwrap();
    assert_eq!(result, slot.slot)
}
