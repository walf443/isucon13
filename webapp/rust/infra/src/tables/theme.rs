use sqipe::{col, table, Col, TableRef};

pub struct ThemeTable;

pub const TABLE_THEMES: ThemeTable = ThemeTable;

impl ThemeTable {
    pub fn table_name(&self) -> &'static str {
        "themes"
    }

    pub fn table(&self) -> TableRef {
        table(self.table_name())
    }

    pub fn id(&self) -> Col {
        col("id")
    }

    pub fn user_id(&self) -> Col {
        col("user_id")
    }

    pub fn dark_mode(&self) -> Col {
        col("dark_mode")
    }
}
