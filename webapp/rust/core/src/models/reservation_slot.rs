use fake::Dummy;

#[derive(Debug, Dummy)]
pub struct ReservationSlot {
    #[allow(unused)]
    pub id: ReservationSlotId,
    pub slot: i64,
    pub start_at: i64,
    pub end_at: i64,
}

kubetsu::define_id!(
    #[derive(toasty::Embed)]
    pub struct ReservationSlotId(i64);
);
kubetsu_serde::impl_serde!(ReservationSlotId(i64));
kubetsu_fake::impl_fake!(ReservationSlotId(i64));
