use isupipe_core::models::user::UserId;

qbey::qbey_schema!(IconTable, "icons", [id: i64, user_id: UserId, image: Vec<u8>], row = IconRow);

pub const TABLE_ICONS: IconTable = IconTable::new();
