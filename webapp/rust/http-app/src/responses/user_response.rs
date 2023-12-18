use isupipe_core::db::DBConn;
use isupipe_core::models::user::UserModel;
use isupipe_core::repos::icon_repository::IconRepository;
use isupipe_core::repos::theme_repository::ThemeRepository;
use isupipe_http_core::responses::ResponseResult;
use isupipe_http_core::FALLBACK_IMAGE;
use isupipe_infra::repos::icon_repository::IconRepositoryInfra;
use isupipe_infra::repos::theme_repository::ThemeRepositoryInfra;
use crate::responses::theme_response::ThemeResponse;

#[derive(Debug, serde::Serialize)]
pub struct UserResponse {
    pub id: i64,
    pub name: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub display_name: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub description: Option<String>,
    pub theme: ThemeResponse,
    pub icon_hash: String,
}

impl UserResponse {
    pub async fn build(conn: &mut DBConn, user: UserModel) -> ResponseResult<Self> {
        let theme_repo = ThemeRepositoryInfra {};
        let theme_model = theme_repo.find_by_user_id(conn, user.id).await?;

        let icon_repo = IconRepositoryInfra {};
        let image = icon_repo.find_image_by_user_id(conn, user.id).await?;

        let image = if let Some(image) = image {
            image
        } else {
            tokio::fs::read(FALLBACK_IMAGE).await?
        };
        use sha2::digest::Digest as _;
        let icon_hash = sha2::Sha256::digest(image);

        Ok(Self {
            id: user.id,
            name: user.name,
            display_name: user.display_name,
            description: user.description,
            theme: ThemeResponse {
                id: theme_model.id,
                dark_mode: theme_model.dark_mode,
            },
            icon_hash: format!("{:x}", icon_hash),
        })
    }
}
