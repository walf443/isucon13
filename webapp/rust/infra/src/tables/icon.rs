use isupipe_core::models::icon::IconId;
use isupipe_core::models::user::UserId;

#[derive(Debug, toasty::Model)]
#[table = "icons"]
pub struct IconRow {
    #[key]
    #[auto]
    pub id: IconId,
    pub user_id: UserId,
    pub image: Vec<u8>,
}
