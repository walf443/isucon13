use sqipe::{col, table, Col, TableRef};

pub struct NgWordTable;

pub const TABLE_NG_WORDS: NgWordTable = NgWordTable;

impl NgWordTable {
    pub fn table_name(&self) -> &'static str {
        "ng_words"
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

    pub fn livestream_id(&self) -> Col {
        col("livestream_id")
    }

    pub fn word(&self) -> Col {
        col("word")
    }

    pub fn created_at(&self) -> Col {
        col("created_at")
    }
}
