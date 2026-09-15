use crate::models::user::UserId;

#[derive(fake::Dummy)]
pub struct CreateIcon {
    pub user_id: UserId,
    pub image: Vec<u8>,
}
