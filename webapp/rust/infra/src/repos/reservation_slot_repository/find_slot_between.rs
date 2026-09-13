use crate::repos::reservation_slot_repository::ReservationSlotRepositoryInfra;
use crate::test_support::InsertReservationSlotSetup;
use crate::test_support::get_db_pool;
use fake::{Fake, Faker};
use isupipe_core::models::reservation_slot::ReservationSlot;
use isupipe_core::repos::reservation_slot_repository::ReservationSlotRepository;

#[tokio::test]
async fn empty_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let repo = ReservationSlotRepositoryInfra {};
    let start_at: i64 = Faker.fake();
    let end_at: i64 = Faker.fake();

    let result = repo.find_slot_between(&mut tx, start_at, end_at).await;
    assert!(result.is_err());
    let err_msg = format!("{}", result.unwrap_err());
    assert!(
        err_msg.contains("record not found"),
        "expected record not found, got: {}",
        err_msg
    );
}

#[tokio::test]
async fn success_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let repo = ReservationSlotRepositoryInfra {};

    let slot: ReservationSlot = Faker.fake();
    {
        InsertReservationSlotSetup { slot: &slot }
            .insert(&mut tx)
            .await;
    }

    let result = repo
        .find_slot_between(&mut tx, slot.start_at, slot.end_at)
        .await
        .unwrap();
    assert_eq!(result, slot.slot)
}
