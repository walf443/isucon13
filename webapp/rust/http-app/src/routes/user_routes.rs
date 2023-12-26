use crate::responses::livestream_response::LivestreamResponse;
use crate::responses::theme_response::ThemeResponse;
use crate::responses::user_response::UserResponse;
use crate::routes::user_icon_routes::get_icon_handler;
use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, State};
use axum::routing::get;
use axum::Router;
use axum_extra::extract::SignedCookieJar;
use isupipe_core::models::user_ranking_entry::UserRankingEntry;
use isupipe_core::models::user_statistics::UserStatistics;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;
use isupipe_core::repos::livestream_repository::LivestreamRepository;
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;
use isupipe_core::repos::reaction_repository::ReactionRepository;
use isupipe_core::repos::theme_repository::ThemeRepository;
use isupipe_core::repos::user_repository::UserRepository;
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};
use isupipe_infra::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use isupipe_infra::repos::livestream_repository::LivestreamRepositoryInfra;
use isupipe_infra::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepositoryInfra;
use isupipe_infra::repos::reaction_repository::ReactionRepositoryInfra;
use isupipe_infra::repos::theme_repository::ThemeRepositoryInfra;
use isupipe_infra::repos::user_repository::UserRepositoryInfra;

pub fn user_routes() -> Router<AppState> {
    let user_routes = Router::new()
        .route("/me", axum::routing::get(get_me_handler))
        // フロントエンドで、配信予約のコラボレーターを指定する際に必要
        .route("/:username", axum::routing::get(get_user_handler))
        .route("/:username/theme", get(get_streamer_theme_handler))
        .route("/:username/livestream", get(get_user_livestreams_handler))
        .route(
            "/:username/statistics",
            axum::routing::get(get_user_statistics_handler),
        )
        .route("/:username/icon", axum::routing::get(get_icon_handler));

    user_routes
}

// 配信者のテーマ取得API
// GET /api/user/:username/theme
pub async fn get_streamer_theme_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<ThemeResponse>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let user_repo = UserRepositoryInfra {};
    let user_id = user_repo
        .find_id_by_name(&mut *tx, &username)
        .await?
        .ok_or(Error::NotFound(
            "not found user that has the given username".into(),
        ))?;

    let theme_repo = ThemeRepositoryInfra {};
    let theme_model = theme_repo.find_by_user_id(&mut *tx, user_id).await?;

    tx.commit().await?;

    Ok(axum::Json(ThemeResponse {
        id: theme_model.id.get(),
        dark_mode: theme_model.dark_mode,
    }))
}
pub async fn get_user_livestreams_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<Vec<LivestreamResponse>>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;
    let user_repo = UserRepositoryInfra {};

    let user = user_repo
        .find_by_name(&mut *tx, &username)
        .await?
        .ok_or(Error::NotFound("user not found".into()))?;

    let livestream_repo = LivestreamRepositoryInfra {};
    let livestream_models = livestream_repo
        .find_all_by_user_id(&mut *tx, &user.id)
        .await?;

    let mut livestreams = Vec::with_capacity(livestream_models.len());
    for livestream_model in livestream_models {
        let livestream = LivestreamResponse::build(&mut tx, livestream_model).await?;
        livestreams.push(livestream);
    }

    tx.commit().await?;

    Ok(axum::Json(livestreams))
}
pub async fn get_me_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
) -> Result<axum::Json<UserResponse>, Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let mut tx = pool.begin().await?;

    let user_repo = UserRepositoryInfra {};
    let user_model = user_repo
        .find(&mut *tx, user_id)
        .await?
        .ok_or(Error::NotFound(
            "not found user that has the userid in session".into(),
        ))?;

    let user = UserResponse::build(&mut *tx, user_model).await?;

    tx.commit().await?;

    Ok(axum::Json(user))
}

// ユーザ詳細API
// GET /api/user/:username
pub async fn get_user_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((username,)): Path<(String,)>,
) -> Result<axum::Json<UserResponse>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let user_repo = UserRepositoryInfra {};
    let user_model = user_repo
        .find_by_name(&mut *tx, &username)
        .await?
        .ok_or(Error::NotFound(
            "not found user that has the given username".into(),
        ))?;

    let user = UserResponse::build(&mut *tx, user_model).await?;

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
    let users = user_repo.find_all(&mut *tx).await?;

    let mut ranking = Vec::new();
    let comment_repo = LivestreamCommentRepositoryInfra {};
    let reaction_repo = ReactionRepositoryInfra {};
    for user in users {
        let reaction_count = reaction_repo
            .count_by_livestream_user_id(&mut *tx, &user.id)
            .await?;

        let tips = comment_repo
            .get_sum_tip_of_livestream_user_id(&mut *tx, &user.id)
            .await?;

        let score = reaction_count + tips;
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
    let total_reactions = reaction_repo
        .count_by_livestream_user_name(&mut *tx, &username)
        .await?;

    // ライブコメント数、チップ合計
    let mut total_livecomments = 0;
    let mut total_tip = 0;

    let livestream_repo = LivestreamRepositoryInfra {};
    let livestreams = livestream_repo
        .find_all_by_user_id(&mut *tx, &user.id)
        .await?;

    let comment_repo = LivestreamCommentRepositoryInfra {};
    for livestream in &livestreams {
        let comments = comment_repo
            .find_all_by_livestream_id(&mut *tx, &livestream.id)
            .await?;

        for comment in comments {
            total_tip += comment.tip;
            total_livecomments += 1;
        }
    }

    let history_repo = LivestreamViewersHistoryRepositoryInfra {};
    let mut conn = pool.acquire().await?;
    // 合計視聴者数
    let mut viewers_count = 0;
    for livestream in livestreams {
        let cnt = history_repo
            .count_by_livestream_id(&mut conn, &livestream.id)
            .await?;
        viewers_count += cnt;
    }

    // お気に入り絵文字
    let favorite_emoji = reaction_repo
        .most_favorite_emoji_by_livestream_user_name(&mut *tx, &username)
        .await?;

    Ok(axum::Json(UserStatistics {
        rank,
        viewers_count,
        total_reactions,
        total_livecomments,
        total_tip,
        favorite_emoji,
    }))
}
