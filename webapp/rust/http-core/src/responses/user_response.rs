use crate::responses::theme_response::ThemeResponse;
use crate::responses::ResponseResult;
use crate::FALLBACK_IMAGE;
use isupipe_core::models::user::User;
use isupipe_core::services::icon_service::IconService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::theme_service::ThemeService;

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
    pub async fn build_by_service<S: ServiceManager>(
        service: &S,
        user: &User,
    ) -> ResponseResult<Self> {
        let theme_model = service.theme_service().find_by_user_id(&user.id).await?;

        let image = service
            .icon_service()
            .find_image_by_user_id(&user.id)
            .await?;

        let image = if let Some(image) = image {
            image
        } else {
            tokio::fs::read(FALLBACK_IMAGE).await?
        };

        use sha2::digest::Digest as _;
        let icon_hash = sha2::Sha256::digest(image);

        Ok(Self {
            id: user.id.get(),
            name: user.name.clone(),
            display_name: user.display_name.clone(),
            description: user.description.clone(),
            theme: ThemeResponse {
                id: theme_model.id.get(),
                dark_mode: theme_model.dark_mode,
            },
            icon_hash: format!("{:x}", icon_hash),
        })
    }
}
