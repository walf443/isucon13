use axum::extract::State;
use axum::http::StatusCode;
use isupipe_core::models::user::CreateUser;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::user_service::UserService;
use crate::error::Error;
use crate::responses::user_response::UserResponse;
use crate::state::AppState;

#[derive(Debug, serde::Deserialize)]
pub struct PostUserRequest {
    name: String,
    display_name: String,
    description: String,
    // password is non-hashed password.
    password: String,
    theme: PostUserRequestTheme,
}

#[derive(Debug, serde::Deserialize)]
pub struct PostUserRequestTheme {
    dark_mode: bool,
}

// ユーザ登録API
// POST /api/register
pub async fn register_handler<S: ServiceManager>(
    State(AppState {
        service,
        powerdns_subdomain_address,
        ..
    }): State<AppState<S>>,
    axum::Json(req): axum::Json<PostUserRequest>,
) -> Result<(StatusCode, axum::Json<UserResponse>), Error> {
    if req.name == "pipe" {
        return Err(Error::BadRequest("the username 'pipe' is reserved".into()));
    }

    let (user, output) = service
        .user_service()
        .create(
            &CreateUser {
                name: req.name.clone(),
                display_name: req.display_name.clone(),
                description: req.description.clone(),
                password: req.password.clone(),
            },
            req.theme.dark_mode,
            &*powerdns_subdomain_address,
        )
        .await?;

    if !output.success {
        return Err(Error::InternalServerError(format!(
            "pdnsutil failed with stdout={} stderr={}",
            String::from_utf8_lossy(&output.stdout),
            String::from_utf8_lossy(&output.stderr),
        )));
    }

    let user = UserResponse::build_by_service(&service, &user).await?;

    Ok((StatusCode::CREATED, axum::Json(user)))
}
