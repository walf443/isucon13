use crate::utils::{fill_livestream_response, fill_user_response};
use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, State};
use axum_extra::extract::SignedCookieJar;
use isupipe_core::models::livestream::{Livestream, LivestreamModel};
use isupipe_core::models::livestream_comment::LivestreamCommentModel;
use isupipe_core::models::mysql_decimal::MysqlDecimal;
use isupipe_core::models::theme::{Theme, ThemeModel};
use isupipe_core::models::user::{User, UserModel};
use isupipe_core::models::user_ranking_entry::UserRankingEntry;
use isupipe_core::models::user_statistics::UserStatistics;
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;
use isupipe_core::repos::user_repository::UserRepository;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};
use isupipe_infra::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepositoryInfra;
use isupipe_infra::repos::user_repository::UserRepositoryInfra;

// 配信者のテーマ取得API
// GET /api/user/:username/theme
pub async fn get_streamer_theme_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<Theme>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let user_repo = UserRepositoryInfra {};
    let user_id = user_repo.find_id_by_name(&mut *tx, &username).await?
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
pub async fn get_user_livestreams_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<Vec<Livestream>>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;
    let user_repo = UserRepositoryInfra {};

    let user = user_repo
        .find_by_name(&mut *tx, &username)
        .await?
        .ok_or(Error::NotFound("user not found".into()))?;

    let livestream_models: Vec<LivestreamModel> =
        sqlx::query_as("SELECT * FROM livestreams WHERE user_id = ?")
            .bind(user.id)
            .fetch_all(&mut *tx)
            .await?;
    let mut livestreams = Vec::with_capacity(livestream_models.len());
    for livestream_model in livestream_models {
        let livestream = fill_livestream_response(&mut tx, livestream_model).await?;
        livestreams.push(livestream);
    }

    tx.commit().await?;

    Ok(axum::Json(livestreams))
}
pub async fn get_me_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
) -> Result<axum::Json<User>, Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let mut tx = pool.begin().await?;

    let user_repo = UserRepositoryInfra {};
    let user_model = user_repo.find(&mut *tx, user_id).await?
        .ok_or(Error::NotFound(
            "not found user that has the userid in session".into(),
        ))?;

    let user = fill_user_response(&mut tx, user_model).await?;

    tx.commit().await?;

    Ok(axum::Json(user))
}

// ユーザ詳細API
// GET /api/user/:username
pub async fn get_user_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<User>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let user_repo = UserRepositoryInfra {};
    let user_model = user_repo
        .find_by_name(&mut *tx, &username)
        .await?
        .ok_or(Error::NotFound(
            "not found user that has the given username".into(),
        ))?;

    let user = fill_user_response(&mut tx, user_model).await?;

    tx.commit().await?;

    Ok(axum::Json(user))
}

pub async fn get_user_statistics_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<UserStatistics>, Error> {
    verify_user_session(&jar).await?;

    // ユーザごとに、紐づく配信について、累計リアクション数、累計ライブコメント数、累計売上金額を算出
    // また、現在の合計視聴者数もだす

    let mut tx = pool.begin().await?;

    let user_repo = UserRepositoryInfra {};
    let user = user_repo
        .find_by_name(&mut *tx, &username)
        .await?
        .ok_or(Error::BadRequest("".into()))?;

    // ランク算出
    let users: Vec<UserModel> = sqlx::query_as("SELECT * FROM users")
        .fetch_all(&mut *tx)
        .await?;

    let mut ranking = Vec::new();
    for user in users {
        let query = r#"
        SELECT COUNT(*) FROM users u
        INNER JOIN livestreams l ON l.user_id = u.id
        INNER JOIN reactions r ON r.livestream_id = l.id
        WHERE u.id = ?
        "#;
        let MysqlDecimal(reactions) = sqlx::query_scalar(query)
            .bind(user.id)
            .fetch_one(&mut *tx)
            .await?;

        let query = r#"
        SELECT IFNULL(SUM(l2.tip), 0) FROM users u
        INNER JOIN livestreams l ON l.user_id = u.id
        INNER JOIN livecomments l2 ON l2.livestream_id = l.id
        WHERE u.id = ?
        "#;
        let MysqlDecimal(tips) = sqlx::query_scalar(query)
            .bind(user.id)
            .fetch_one(&mut *tx)
            .await?;

        let score = reactions + tips;
        ranking.push(UserRankingEntry {
            username: user.name,
            score,
        });
    }
    ranking.sort_by(|a, b| {
        a.score
            .cmp(&b.score)
            .then_with(|| a.username.cmp(&b.username))
    });

    let rpos = ranking
        .iter()
        .rposition(|entry| entry.username == username)
        .unwrap();
    let rank = (ranking.len() - rpos) as i64;

    // リアクション数
    let query = r"#
    SELECT COUNT(*) FROM users u
    INNER JOIN livestreams l ON l.user_id = u.id
    INNER JOIN reactions r ON r.livestream_id = l.id
    WHERE u.name = ?
    #";
    let MysqlDecimal(total_reactions) = sqlx::query_scalar(query)
        .bind(&username)
        .fetch_one(&mut *tx)
        .await?;

    // ライブコメント数、チップ合計
    let mut total_livecomments = 0;
    let mut total_tip = 0;
    let livestreams: Vec<LivestreamModel> =
        sqlx::query_as("SELECT * FROM livestreams WHERE user_id = ?")
            .bind(user.id)
            .fetch_all(&mut *tx)
            .await?;

    for livestream in &livestreams {
        let livecomments: Vec<LivestreamCommentModel> =
            sqlx::query_as("SELECT * FROM livecomments WHERE livestream_id = ?")
                .bind(livestream.id)
                .fetch_all(&mut *tx)
                .await?;

        for livecomment in livecomments {
            total_tip += livecomment.tip;
            total_livecomments += 1;
        }
    }

    let history_repo = LivestreamViewersHistoryRepositoryInfra {};
    let mut conn = pool.acquire().await?;
    // 合計視聴者数
    let mut viewers_count = 0;
    for livestream in livestreams {
        let cnt = history_repo
            .count_by_livestream_id(&mut conn, livestream.id)
            .await?;
        viewers_count += cnt;
    }

    // お気に入り絵文字
    let query = r#"
    SELECT r.emoji_name
    FROM users u
    INNER JOIN livestreams l ON l.user_id = u.id
    INNER JOIN reactions r ON r.livestream_id = l.id
    WHERE u.name = ?
    GROUP BY emoji_name
    ORDER BY COUNT(*) DESC, emoji_name DESC
    LIMIT 1
    "#;
    let favorite_emoji: String = sqlx::query_scalar(query)
        .bind(&username)
        .fetch_optional(&mut *tx)
        .await?
        .unwrap_or_default();

    Ok(axum::Json(UserStatistics {
        rank,
        viewers_count,
        total_reactions,
        total_livecomments,
        total_tip,
        favorite_emoji,
    }))
}
