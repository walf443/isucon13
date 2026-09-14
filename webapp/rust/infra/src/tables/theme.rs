use isupipe_core::models::theme::{Theme, ThemeId};
use isupipe_core::models::user::UserId;

#[derive(Debug, toasty::Model)]
#[table = "themes"]
pub struct ThemeRow {
    #[key]
    #[auto]
    pub id: ThemeId,
    pub user_id: UserId,
    pub dark_mode: bool,
}

impl From<ThemeRow> for Theme {
    fn from(row: ThemeRow) -> Self {
        Self {
            id: row.id,
            user_id: row.user_id,
            dark_mode: row.dark_mode,
        }
    }
}
