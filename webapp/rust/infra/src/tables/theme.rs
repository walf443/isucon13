use isupipe_core::models::theme::ThemeId;
use isupipe_core::models::user::UserId;

qbey::qbey_schema!(ThemeTable, "themes", [id: ThemeId, user_id: UserId, dark_mode: bool]);

pub const TABLE_THEMES: ThemeTable = ThemeTable::new();
