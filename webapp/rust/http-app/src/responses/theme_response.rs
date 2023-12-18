#[derive(Debug, serde::Deserialize, serde::Serialize)]
pub struct ThemeResponse {
    pub id: i64,
    pub dark_mode: bool,
}
