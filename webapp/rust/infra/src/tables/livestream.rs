use sqipe::{table, Col, TableRef};

pub struct LivestreamTable {
    alias: Option<&'static str>,
}

pub const TABLE_LIVESTREAMS: LivestreamTable = LivestreamTable { alias: None };

impl LivestreamTable {
    pub fn table_name(&self) -> &'static str {
        "livestreams"
    }

    pub fn table(&self) -> TableRef {
        table(self.alias.unwrap_or(self.table_name()))
    }

    pub fn as_(&self, alias: &'static str) -> Self {
        LivestreamTable { alias: Some(alias) }
    }

    pub fn id(&self) -> Col {
        self.table().col("id")
    }

    pub fn user_id(&self) -> Col {
        self.table().col("user_id")
    }
}
