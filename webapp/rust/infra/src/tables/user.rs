qbey::qbey_schema!(UserTable, "users", [
    id,
    name,
    display_name,
    description,
    password,
]);

impl UserTable {
    pub fn default_cols(&self) -> Vec<qbey::Col> {
        vec![
            self.id(),
            self.name(),
            self.display_name(),
            self.description(),
            self.password().as_("hashed_password"),
        ]
    }
}

pub const TABLE_USERS: UserTable = UserTable::new();
