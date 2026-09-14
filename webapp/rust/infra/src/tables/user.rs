use isupipe_core::models::user::{User, UserId, UserName};

#[derive(Debug, toasty::Model)]
#[table = "users"]
pub struct UserRow {
    #[key]
    #[auto]
    pub id: UserId,
    #[unique]
    pub name: UserName,
    pub display_name: String,
    #[column("password")]
    pub hashed_password: String,
    pub description: String,
}

impl From<UserRow> for User {
    fn from(row: UserRow) -> Self {
        Self {
            id: row.id,
            name: row.name,
            display_name: Some(row.display_name),
            description: Some(row.description),
            hashed_password: row.hashed_password,
        }
    }
}
