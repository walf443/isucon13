use sqipe::{col, Col};

pub struct UserTable;

pub const TABLE_USERS: UserTable = UserTable;

impl UserTable {
    pub fn table_name(&self) -> &'static str {
        "users"
    }

    pub fn id(&self) -> Col {
        col("id")
    }

    pub fn name(&self) -> Col {
        col("name")
    }

    pub fn display_name(&self) -> Col {
        col("display_name")
    }

    pub fn description(&self) -> Col {
        col("description")
    }

    pub fn password(&self) -> Col {
        col("password")
    }
}
