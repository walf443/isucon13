#[derive(Debug, toasty::Model)]
#[table = "icons"]
pub struct IconRow {
    #[key]
    #[auto]
    pub id: i64,
    pub user_id: i64,
    pub image: Vec<u8>,
}
