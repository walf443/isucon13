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

    let result = repo
        .find_all_between_for_update(&mut tx, slot1.start_at, slot2.end_at)
        .await
        .unwrap();
    assert_eq!(result.len(), 2);
    let sid = slot1.id.get();
    let got1 = result.iter().find(|i| i.id.get() == sid).unwrap();
    assert_eq!(got1.slot, slot1.slot);
    assert_eq!(got1.start_at, slot1.start_at);
    assert_eq!(got1.end_at, slot1.end_at);

    let sid = slot2.id.get();
    let got2 = result.iter().find(|i| i.id.get() == sid).unwrap();
    assert_eq!(got2.slot, slot2.slot);
    assert_eq!(got2.start_at, slot2.start_at);
    assert_eq!(got2.end_at, slot2.end_at);
}
