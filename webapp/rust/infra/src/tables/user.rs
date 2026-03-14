sqipe_schema!(UserTable, "users", [
    id,
    name,
    display_name,
    description,
    password,
]);

pub const TABLE_USERS: UserTable = UserTable::new();
