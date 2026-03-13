use crate::models::user::UserId;

pub const TABLE_NAME: &str = "icons";

pub struct CreateIcon {
    pub user_id: UserId,
    pub image: Vec<u8>,
}
