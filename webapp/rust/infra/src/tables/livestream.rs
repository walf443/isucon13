use sqipe::{col, Col};

pub struct LivestreamTable;

pub const TABLE_LIVESTREAMS: LivestreamTable = LivestreamTable;

impl LivestreamTable {
    pub fn table_name(&self) -> &'static str {
        "livestreams"
    }

    pub fn id(&self) -> Col {
        col("id")
    }

    pub fn user_id(&self) -> Col {
        col("user_id")
    }
}
