use isupipe_core::models::theme::{Theme, ThemeId};
use isupipe_core::models::user::UserId;

#[derive(Debug, toasty::Model)]
#[table = "themes"]
pub struct ThemeRow {
    #[key]
    #[auto]
    pub id: i64,
    pub user_id: i64,
    pub dark_mode: bool,
}

impl From<ThemeRow> for Theme {
    fn from(row: ThemeRow) -> Self {
        Self {
            id: ThemeId::new(row.id),
            user_id: UserId::new(row.user_id),
            dark_mode: row.dark_mode,
        }
    }
}
