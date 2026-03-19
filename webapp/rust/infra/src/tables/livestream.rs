qbey::qbey_schema!(LivestreamTable, "livestreams", [id, user_id]);

pub const TABLE_LIVESTREAMS: LivestreamTable = LivestreamTable::new();
