use crate::repos::reservation_slot_repository::ReservationSlotRepositoryInfra;
use crate::tables::reservation_slot::ReservationSlotRow;
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
    let slot: ReservationSlot = Faker.fake();
    repo.decrement_slot_between(&mut tx, slot.start_at, slot.end_at)
        .await
        .unwrap();
}

#[tokio::test]
async fn not_empty_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    // deadlock対策
    toasty::sql::query("SELECT id FROM reservation_slots FOR UPDATE")
        .exec(&mut tx)
        .await
        .unwrap();

    let repo = ReservationSlotRepositoryInfra {};
    let mut slot1: ReservationSlot = Faker.fake();
    let mut slot2: ReservationSlot = Faker.fake();
    slot1.end_at = slot1.start_at + 100;
    slot2.start_at = slot1.end_at + 100;
    slot2.end_at = slot2.start_at + 100;

    {
        InsertReservationSlotSetup { slot: &slot1 }
            .insert(&mut tx)
            .await;
        InsertReservationSlotSetup { slot: &slot2 }
            .insert(&mut tx)
            .await;
    }

    repo.decrement_slot_between(&mut tx, slot1.start_at, slot2.end_at)
        .await
        .unwrap();

    let got1: ReservationSlot =
        ReservationSlotRow::filter(ReservationSlotRow::fields().id().eq(slot1.id.inner()))
            .one()
            .exec(&mut tx)
            .await
            .unwrap()
            .into();
    assert_eq!(slot1.slot - 1, got1.slot);

    let got2: ReservationSlot =
        ReservationSlotRow::filter(ReservationSlotRow::fields().id().eq(slot2.id.inner()))
            .one()
            .exec(&mut tx)
            .await
            .unwrap()
            .into();
    assert_eq!(slot2.slot - 1, got2.slot);
}
