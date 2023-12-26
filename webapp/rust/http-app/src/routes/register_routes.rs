use crate::responses::user_response::UserResponse;
use axum::extract::State;
use axum::http::StatusCode;
use isupipe_core::models::user::User;
use isupipe_core::repos::theme_repository::ThemeRepository;
use isupipe_core::repos::user_repository::UserRepository;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_infra::repos::theme_repository::ThemeRepositoryInfra;
use isupipe_infra::repos::user_repository::UserRepositoryInfra;

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
pub async fn register_handler(
    State(AppState {
        pool,
        powerdns_subdomain_address,
        ..
    }): State<AppState>,
    axum::Json(req): axum::Json<PostUserRequest>,
) -> Result<(StatusCode, axum::Json<UserResponse>), Error> {
    if req.name == "pipe" {
        return Err(Error::BadRequest("the username 'pipe' is reserved".into()));
    }

    const BCRYPT_DEFAULT_COST: u32 = 4;
    let hashed_password = bcrypt::hash(&req.password, BCRYPT_DEFAULT_COST)?;

    let mut tx = pool.begin().await?;

    let user_repo = UserRepositoryInfra {};
    let user_id = user_repo
        .insert(
            &mut *tx,
            &req.name,
            &req.display_name,
            &req.description,
            &req.password,
        )
        .await?;

    let theme_repo = ThemeRepositoryInfra {};
    theme_repo
        .insert(&mut *tx, &user_id, req.theme.dark_mode)
        .await?;

    let output = tokio::process::Command::new("pdnsutil")
        .arg("add-record")
        .arg("u.isucon.dev")
        .arg(&req.name)
        .arg("A")
        .arg("0")
        .arg(&*powerdns_subdomain_address)
        .output()
        .await?;
    if !output.status.success() {
        return Err(Error::InternalServerError(format!(
            "pdnsutil failed with stdout={} stderr={}",
            String::from_utf8_lossy(&output.stdout),
            String::from_utf8_lossy(&output.stderr),
        )));
    }

    let user = UserResponse::build(
        &mut *tx,
        User {
            id: user_id.clone(),
            name: req.name,
            display_name: Some(req.display_name),
            description: Some(req.description),
            hashed_password: Some(hashed_password),
        },
    )
    .await?;

    tx.commit().await?;

    Ok((StatusCode::CREATED, axum::Json(user)))
}
