use crate::utils::fill_user_response;
use axum::extract::State;
use axum::http::StatusCode;
use isupipe_core::models::user::{User, UserModel};
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;

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
) -> Result<(StatusCode, axum::Json<User>), Error> {
    if req.name == "pipe" {
        return Err(Error::BadRequest("the username 'pipe' is reserved".into()));
    }

    const BCRYPT_DEFAULT_COST: u32 = 4;
    let hashed_password = bcrypt::hash(&req.password, BCRYPT_DEFAULT_COST)?;

    let mut tx = pool.begin().await?;

    let result = sqlx::query(
        "INSERT INTO users (name, display_name, description, password) VALUES(?, ?, ?, ?)",
    )
    .bind(&req.name)
    .bind(&req.display_name)
    .bind(&req.description)
    .bind(&hashed_password)
    .execute(&mut *tx)
    .await?;
    let user_id = result.last_insert_id() as i64;

    sqlx::query("INSERT INTO themes (user_id, dark_mode) VALUES(?, ?)")
        .bind(user_id)
        .bind(req.theme.dark_mode)
        .execute(&mut *tx)
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

    let user = fill_user_response(
        &mut tx,
        UserModel {
            id: user_id,
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
