use crate::db::DBConn;
use crate::models::reservation_slot::ReservationSlot;
use crate::repos::Result;
use async_trait::async_trait;

#[cfg_attr(any(feature = "test", test), mockall::automock)]
#[async_trait]
pub trait ReservationSlotRepository {
    async fn find_all_between_for_update(
        &self,
        conn: &mut DBConn,
        start_at: i64,
        end_at: i64,
    ) -> Result<Vec<ReservationSlot>>;

    async fn find_slot_between(&self, conn: &mut DBConn, start_at: i64, end_at: i64)
        -> Result<i64>;

    async fn decrement_slot_between(
        &self,
        conn: &mut DBConn,
        start_at: i64,
        end_at: i64,
    ) -> Result<()>;
}

pub trait HaveReservationSlotRepository {
    type Repo: Sync + ReservationSlotRepository;

    fn reservation_slot_repo(&self) -> &Self::Repo;
}
