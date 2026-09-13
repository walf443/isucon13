use isupipe_core::models::user::UserId;

#[derive(Debug, toasty::Model)]
#[table = "icons"]
pub struct IconRow {
    #[key]
    #[auto]
    pub id: i64,
    pub user_id: UserId,
    pub image: Vec<u8>,
}
