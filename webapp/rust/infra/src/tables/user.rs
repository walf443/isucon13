use sqipe::{table, Col, TableRef};

pub struct UserTable {
    alias: Option<&'static str>,
}

pub const TABLE_USERS: UserTable = UserTable { alias: None };

impl UserTable {
    pub fn table_name(&self) -> &'static str {
        "users"
    }

    pub fn table(&self) -> TableRef {
        table(self.alias.unwrap_or(self.table_name()))
    }

    pub fn as_(&self, alias: &'static str) -> Self {
        UserTable { alias: Some(alias) }
    }

    pub fn id(&self) -> Col {
        self.table().col("id")
    }

    pub fn name(&self) -> Col {
        self.table().col("name")
    }

    pub fn display_name(&self) -> Col {
        self.table().col("display_name")
    }

    pub fn description(&self) -> Col {
        self.table().col("description")
    }

    pub fn password(&self) -> Col {
        self.table().col("password")
    }
}
