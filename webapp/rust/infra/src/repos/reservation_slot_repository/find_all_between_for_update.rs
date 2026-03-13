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
    let result = repo
        .find_all_between_for_update(&mut tx, slot.start_at, slot.end_at)
        .await
        .unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn not_empty_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = ReservationSlotRepositoryInfra {};
    let mut slot1: ReservationSlot = Faker.fake();
    slot1.end_at = slot1.start_at + 100;
    let mut slot2: ReservationSlot = Faker.fake();
    slot2.start_at = slot1.end_at + 100;
    slot2.end_at = slot2.start_at + 100;

    sqlx::query(
        "INSERT INTO reservation_slots (id, slot, start_at, end_at) VALUES (?, ?, ?, ?), (?, ?, ?, ?)",
    )
    .bind(&slot1.id)
    .bind(slot1.slot)
    .bind(slot1.start_at)
    .bind(slot1.end_at)
    .bind(&slot2.id)
    .bind(slot2.slot)
    .bind(slot2.start_at)
    .bind(slot2.end_at)
    .execute(&mut *tx)
    .await
    .unwrap();

    let result = repo
        .find_all_between_for_update(&mut tx, slot1.start_at, slot2.end_at)
        .await
        .unwrap();
    assert_eq!(result.len(), 2);
    let sid = slot1.id.clone();
    let got1 = result.iter().find(|i| i.id == sid).unwrap();
    assert_eq!(got1.slot, slot1.slot);
    assert_eq!(got1.start_at, slot1.start_at);
    assert_eq!(got1.end_at, slot1.end_at);

    let sid = slot2.id.clone();
    let got2 = result.iter().find(|i| i.id == sid).unwrap();
    assert_eq!(got2.slot, slot2.slot);
    assert_eq!(got2.start_at, slot2.start_at);
    assert_eq!(got2.end_at, slot2.end_at);
}

#[tokio::test]
async fn boundary_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = ReservationSlotRepositoryInfra {};
    let mut slot: ReservationSlot = Faker.fake();
    slot.start_at = 1000;
    slot.end_at = 2000;

    sqlx::query(
        "INSERT INTO reservation_slots (id, slot, start_at, end_at) VALUES (?, ?, ?, ?)",
    )
    .bind(&slot.id)
    .bind(slot.slot)
    .bind(slot.start_at)
    .bind(slot.end_at)
    .execute(&mut *tx)
    .await
    .unwrap();

    // exact boundary match (start_at >= 1000 AND end_at <= 2000)
    let result = repo
        .find_all_between_for_update(&mut tx, 1000, 2000)
        .await
        .unwrap();
    assert_eq!(result.len(), 1);

    // start_at is after slot.start_at → excluded
    let result = repo
        .find_all_between_for_update(&mut tx, 1001, 2000)
        .await
        .unwrap();
    assert_eq!(result.len(), 0);

    // end_at is before slot.end_at → excluded
    let result = repo
        .find_all_between_for_update(&mut tx, 1000, 1999)
        .await
        .unwrap();
    assert_eq!(result.len(), 0);
}

#[tokio::test]
async fn filters_out_of_range() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = ReservationSlotRepositoryInfra {};
    let mut inside: ReservationSlot = Faker.fake();
    inside.start_at = 1000;
    inside.end_at = 2000;
    let mut outside: ReservationSlot = Faker.fake();
    outside.start_at = 3000;
    outside.end_at = 4000;

    sqlx::query(
        "INSERT INTO reservation_slots (id, slot, start_at, end_at) VALUES (?, ?, ?, ?), (?, ?, ?, ?)",
    )
    .bind(&inside.id)
    .bind(inside.slot)
    .bind(inside.start_at)
    .bind(inside.end_at)
    .bind(&outside.id)
    .bind(outside.slot)
    .bind(outside.start_at)
    .bind(outside.end_at)
    .execute(&mut *tx)
    .await
    .unwrap();

    let result = repo
        .find_all_between_for_update(&mut tx, 1000, 2000)
        .await
        .unwrap();
    assert_eq!(result.len(), 1);
    assert_eq!(result[0].id, inside.id);
}
