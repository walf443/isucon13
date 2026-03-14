sqipe_schema!(UserTable, "users", [
    id,
    name,
    display_name,
    description,
    password,
]);

impl UserTable {
    pub fn default_cols(&self) -> Vec<sqipe::ColRef> {
        use sqipe::IntoColRef;
        vec![
            self.id().into_col_ref(),
            self.name().into_col_ref(),
            self.display_name().into_col_ref(),
            self.description().into_col_ref(),
            self.password().as_("hashed_password"),
        ]
    }
}

pub const TABLE_USERS: UserTable = UserTable::new();
