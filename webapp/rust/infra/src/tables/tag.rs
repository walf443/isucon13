use sqipe::{table, Col, TableRef};

pub struct TagTable;

pub const TABLE_TAGS: TagTable = TagTable;

impl TagTable {
    pub fn table_name(&self) -> &'static str {
        "tags"
    }

    pub fn table(&self) -> TableRef {
        table(self.table_name())
    }

    pub fn id(&self) -> Col {
        self.table().col("id")
    }

    pub fn name(&self) -> Col {
        self.table().col("name")
    }
}
