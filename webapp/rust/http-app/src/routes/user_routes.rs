use axum::extract::{Path, State};
use axum_extra::extract::SignedCookieJar;
use isupipe_core::models::theme::{Theme, ThemeModel};
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::verify_user_session;

// 配信者のテーマ取得API
// GET /api/user/:username/theme
pub async fn get_streamer_theme_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<Theme>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let user_id: i64 = sqlx::query_scalar("SELECT id FROM users WHERE name = ?")
        .bind(username)
        .fetch_optional(&mut *tx)
        .await?
        .ok_or(Error::NotFound(
            "not found user that has the given username".into(),
        ))?;

    let theme_model: ThemeModel = sqlx::query_as("SELECT * FROM themes WHERE user_id = ?")
        .bind(user_id)
        .fetch_one(&mut *tx)
        .await?;

    tx.commit().await?;

    Ok(axum::Json(Theme {
        id: theme_model.id,
        dark_mode: theme_model.dark_mode,
    }))
}
