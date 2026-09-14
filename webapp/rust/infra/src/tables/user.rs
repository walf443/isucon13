use isupipe_core::models::user::{UserId, UserName};

qbey::qbey_schema!(
    UserTable,
    "users",
    [
        id: UserId,
        name: UserName,
        display_name: String,
        description: String,
        password: String,
    ]
);

impl UserTable {
    pub fn default_cols(&self) -> Vec<qbey::Col> {
        vec![
            self.id().into_col(),
            self.name().into_col(),
            self.display_name().into_col(),
            self.description().into_col(),
            self.password().as_("hashed_password").into_col(),
        ]
    }
}

pub const TABLE_USERS: UserTable = UserTable::new();
