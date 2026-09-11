qbey::qbey_schema!(ThemeTable, "themes", [id, user_id, dark_mode]);

pub const TABLE_THEMES: ThemeTable = ThemeTable::new();
