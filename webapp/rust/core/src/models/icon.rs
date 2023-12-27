use crate::models::user::UserId;

pub struct CreateIcon {
    pub user_id: UserId,
    pub image: Vec<u8>,
}
