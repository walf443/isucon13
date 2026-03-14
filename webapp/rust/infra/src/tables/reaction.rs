use sqipe::{col, Col};

pub struct ReactionTable;

pub const TABLE_REACTIONS: ReactionTable = ReactionTable;

impl ReactionTable {
    pub fn table_name(&self) -> &'static str {
        "reactions"
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

    pub fn created_at(&self) -> Col {
        col("created_at")
    }
}
