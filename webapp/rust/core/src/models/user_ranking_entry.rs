use crate::models::user::UserName;

#[derive(Debug)]
pub struct UserRankingEntry {
    pub username: UserName,
    pub score: i64,
}
