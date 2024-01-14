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
    sqlx::query!("SELECT id FROM reservation_slots FOR UPDATE")
        .fetch_all(&mut *tx)
        .await
        .unwrap();

    let repo = ReservationSlotRepositoryInfra {};
    let mut slot1: ReservationSlot = Faker.fake();
    let mut slot2: ReservationSlot = Faker.fake();
    slot1.end_at = slot1.start_at + 100;
    slot2.start_at = slot1.end_at + 100;
    slot2.end_at = slot2.start_at + 100;

    sqlx::query!(
        "INSERT INTO reservation_slots (id,slot,start_at, end_at) VALUES (?, ?, ?, ?), (?, ?, ?, ?)",
        &slot1.id,
        slot1.slot,
        slot1.start_at,
        slot1.end_at,
        &slot2.id,
        slot2.slot,
        slot2.start_at,
        slot2.end_at,
    ).execute(&mut *tx).await.unwrap();

    repo.decrement_slot_between(&mut tx, slot1.start_at, slot2.end_at)
        .await
        .unwrap();

    let got1 = sqlx::query!("SELECT * FROM reservation_slots WHERE id = ?", slot1.id)
        .fetch_one(&mut *tx)
        .await
        .unwrap();
    assert_eq!(slot1.slot - 1, got1.slot);

    let got2 = sqlx::query!("SELECT * FROM reservation_slots WHERE id = ?", slot2.id)
        .fetch_one(&mut *tx)
        .await
        .unwrap();
    assert_eq!(slot2.slot - 1, got2.slot);
}
