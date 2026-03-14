use sqipe::{table, Col, TableRef};

pub struct IconTable;

pub const TABLE_ICONS: IconTable = IconTable;

impl IconTable {
    pub fn table_name(&self) -> &'static str {
        "icons"
    }

    pub fn table(&self) -> TableRef {
        table(self.table_name())
    }

    pub fn image(&self) -> Col {
        self.table().col("image")
    }

    pub fn user_id(&self) -> Col {
        self.table().col("user_id")
    }
}
