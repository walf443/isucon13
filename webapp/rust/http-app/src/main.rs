use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, State};
use axum::http::StatusCode;
use axum_extra::extract::cookie::SignedCookieJar;
use chrono::Utc;
use isupipe_core::models::livestream::LivestreamModel;
use isupipe_core::models::mysql_decimal::MysqlDecimal;
use isupipe_http_app::routes::initialize_routes::initialize_handler;
use isupipe_http_app::routes::livestream_comment_report_routes::{
    get_livecomment_reports_handler, report_livecomment_handler,
};
use isupipe_http_app::routes::livestream_comment_routes::{
    get_livecomments_handler, post_livecomment_handler,
};
use isupipe_http_app::routes::livestream_reaction_routes::{
    get_reactions_handler, post_reaction_handler,
};
use isupipe_http_app::routes::livestream_routes::{
    enter_livestream_handler, exit_livestream_handler, get_livestream_handler,
    get_my_livestreams_handler, get_ngwords, moderate_handler, reserve_livestream_handler,
    search_livestreams_handler,
};
use isupipe_http_app::routes::login_routes::login_handler;
use isupipe_http_app::routes::register_routes::register_handler;
use isupipe_http_app::routes::tag_routes::get_tag_handler;
use isupipe_http_app::routes::user_routes::{get_icon_handler, get_me_handler, get_streamer_theme_handler, get_user_handler, get_user_livestreams_handler, get_user_statistics_handler};
use isupipe_http_core::error::Error;
use isupipe_http_core::state::AppState;
use std::sync::Arc;

const DEFAULT_SESSION_ID_KEY: &str = "SESSIONID";
const DEFUALT_SESSION_EXPIRES_KEY: &str = "EXPIRES";
const DEFAULT_USER_ID_KEY: &str = "USERID";

fn build_mysql_options() -> sqlx::mysql::MySqlConnectOptions {
    let mut options = sqlx::mysql::MySqlConnectOptions::new()
        .host("127.0.0.1")
        .port(3306)
        .username("isucon")
        .password("isucon")
        .database("isupipe");
    if let Ok(host) = std::env::var("ISUCON13_MYSQL_DIALCONFIG_ADDRESS") {
        options = options.host(&host);
    }
    if let Some(port) = std::env::var("ISUCON13_MYSQL_DIALCONFIG_PORT")
        .ok()
        .and_then(|port_str| port_str.parse().ok())
    {
        options = options.port(port);
    }
    if let Ok(user) = std::env::var("ISUCON13_MYSQL_DIALCONFIG_USER") {
        options = options.username(&user);
    }
    if let Ok(password) = std::env::var("ISUCON13_MYSQL_DIALCONFIG_PASSWORD") {
        options = options.password(&password);
    }
    if let Ok(database) = std::env::var("ISUCON13_MYSQL_DIALCONFIG_DATABASE") {
        options = options.database(&database);
    }
    options
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    if std::env::var_os("RUST_LOG").is_none() {
        std::env::set_var("RUST_LOG", "info,tower_http=debug,axum::rejection=trace");
    }
    tracing_subscriber::fmt::init();

    let pool = sqlx::mysql::MySqlPoolOptions::new()
        .connect_with(build_mysql_options())
        .await
        .expect("failed to connect db");

    const DEFAULT_SECRET: &[u8] = b"isucon13_session_cookiestore_defaultsecret";
    let secret = if let Ok(secret) = std::env::var("ISUCON13_SESSION_SECRETKEY") {
        secret.into_bytes()
    } else {
        DEFAULT_SECRET.to_owned()
    };

    const POWERDNS_SUBDOMAIN_ADDRESS_ENV_KEY: &str = "ISUCON13_POWERDNS_SUBDOMAIN_ADDRESS";
    let Ok(powerdns_subdomain_address) = std::env::var(POWERDNS_SUBDOMAIN_ADDRESS_ENV_KEY) else {
        panic!(
            "environ {} must be provided",
            POWERDNS_SUBDOMAIN_ADDRESS_ENV_KEY
        );
    };

    let app = axum::Router::new()
        // 初期化
        .route("/api/initialize", axum::routing::post(initialize_handler))
        // top
        .route("/api/tag", axum::routing::get(get_tag_handler))
        .route(
            "/api/user/:username/theme",
            axum::routing::get(get_streamer_theme_handler),
        )
        // livestream
        // reserve livestream
        .route(
            "/api/livestream/reservation",
            axum::routing::post(reserve_livestream_handler),
        )
        // list livestream
        .route(
            "/api/livestream/search",
            axum::routing::get(search_livestreams_handler),
        )
        .route(
            "/api/livestream",
            axum::routing::get(get_my_livestreams_handler),
        )
        .route(
            "/api/user/:username/livestream",
            axum::routing::get(get_user_livestreams_handler),
        )
        // get livestream
        .route(
            "/api/livestream/:livestream_id",
            axum::routing::get(get_livestream_handler),
        )
        // get polling livecomment timeline
        // ライブコメント投稿
        .route(
            "/api/livestream/:livestream_id/livecomment",
            axum::routing::get(get_livecomments_handler).post(post_livecomment_handler),
        )
        .route(
            "/api/livestream/:livestream_id/reaction",
            axum::routing::get(get_reactions_handler).post(post_reaction_handler),
        )
        // (配信者向け)ライブコメントの報告一覧取得API
        .route(
            "/api/livestream/:livestream_id/report",
            axum::routing::get(get_livecomment_reports_handler),
        )
        .route(
            "/api/livestream/:livestream_id/ngwords",
            axum::routing::get(get_ngwords),
        )
        // ライブコメント報告
        .route(
            "/api/livestream/:livestream_id/livecomment/:livecomment_id/report",
            axum::routing::post(report_livecomment_handler),
        )
        // 配信者によるモデレーション (NGワード登録)
        .route(
            "/api/livestream/:livestream_id/moderate",
            axum::routing::post(moderate_handler),
        )
        // livestream_viewersにINSERTするため必要
        // ユーザ視聴開始 (viewer)
        .route(
            "/api/livestream/:livestream_id/enter",
            axum::routing::post(enter_livestream_handler),
        )
        // ユーザ視聴終了 (viewer)
        .route(
            "/api/livestream/:livestream_id/exit",
            axum::routing::delete(exit_livestream_handler),
        )
        // user
        .route("/api/register", axum::routing::post(register_handler))
        .route("/api/login", axum::routing::post(login_handler))
        .route("/api/user/me", axum::routing::get(get_me_handler))
        // フロントエンドで、配信予約のコラボレーターを指定する際に必要
        .route("/api/user/:username", axum::routing::get(get_user_handler))
        .route(
            "/api/user/:username/statistics",
            axum::routing::get(get_user_statistics_handler),
        )
        .route(
            "/api/user/:username/icon",
            axum::routing::get(get_icon_handler),
        )
        .route("/api/icon", axum::routing::post(post_icon_handler))
        // stats
        // ライブ配信統計情報
        .route(
            "/api/livestream/:livestream_id/statistics",
            axum::routing::get(get_livestream_statistics_handler),
        )
        // 課金情報
        .route("/api/payment", axum::routing::get(get_payment_result))
        .with_state(AppState {
            pool,
            key: axum_extra::extract::cookie::Key::derive_from(&secret),
            powerdns_subdomain_address: Arc::new(powerdns_subdomain_address),
        })
        .layer(tower_http::trace::TraceLayer::new_for_http());

    // HTTPサーバ起動
    if let Some(tcp_listener) = listenfd::ListenFd::from_env().take_tcp_listener(0)? {
        axum::Server::from_tcp(tcp_listener)?
    } else {
        const LISTEN_PORT: u16 = 8080;
        axum::Server::bind(&std::net::SocketAddr::from(([0, 0, 0, 0], LISTEN_PORT)))
    }
    .serve(app.into_make_service())
    .await?;

    Ok(())
}

#[derive(Debug, serde::Deserialize)]
struct PostIconRequest {
    #[serde(deserialize_with = "from_base64")]
    image: Vec<u8>,
}
fn from_base64<'de, D>(deserializer: D) -> Result<Vec<u8>, D::Error>
where
    D: serde::Deserializer<'de>,
{
    use base64::Engine as _;
    use serde::de::{Deserialize as _, Error as _};
    let value = String::deserialize(deserializer)?;
    base64::engine::general_purpose::STANDARD
        .decode(value)
        .map_err(D::Error::custom)
}

#[derive(Debug, serde::Serialize)]
struct PostIconResponse {
    id: i64,
}

async fn post_icon_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    axum::Json(req): axum::Json<PostIconRequest>,
) -> Result<(StatusCode, axum::Json<PostIconResponse>), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;

    let mut tx = pool.begin().await?;

    sqlx::query("DELETE FROM icons WHERE user_id = ?")
        .bind(user_id)
        .execute(&mut *tx)
        .await?;

    let rs = sqlx::query("INSERT INTO icons (user_id, image) VALUES (?, ?)")
        .bind(user_id)
        .bind(req.image)
        .execute(&mut *tx)
        .await?;
    let icon_id = rs.last_insert_id() as i64;

    tx.commit().await?;

    Ok((
        StatusCode::CREATED,
        axum::Json(PostIconResponse { id: icon_id }),
    ))
}

#[derive(Debug, serde::Serialize)]
struct Session {
    id: String,
    user_id: i64,
    expires: i64,
}

async fn verify_user_session(jar: &SignedCookieJar) -> Result<(), Error> {
    let cookie = jar
        .get(DEFAULT_SESSION_ID_KEY)
        .ok_or(Error::Forbidden("".into()))?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::Forbidden("".into()))?;
    let session_expires: i64 = sess
        .get(DEFUALT_SESSION_EXPIRES_KEY)
        .ok_or(Error::Forbidden("".into()))?;
    let now = Utc::now();
    if now.timestamp() > session_expires {
        return Err(Error::Unauthorized("session has expired".into()));
    }
    Ok(())
}

#[derive(Debug, serde::Serialize)]
struct LivestreamStatistics {
    rank: i64,
    viewers_count: i64,
    total_reactions: i64,
    total_reports: i64,
    max_tip: i64,
}

#[derive(Debug)]
struct LivestreamRankingEntry {
    livestream_id: i64,
    score: i64,
}

async fn get_livestream_statistics_handler(
    State(AppState { pool, .. }): State<AppState>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<axum::Json<LivestreamStatistics>, Error> {
    verify_user_session(&jar).await?;

    let mut tx = pool.begin().await?;

    let _: LivestreamModel = sqlx::query_as("SELECT * FROM livestreams WHERE id = ?")
        .bind(livestream_id)
        .fetch_optional(&mut *tx)
        .await?
        .ok_or(Error::BadRequest("".into()))?;

    let livestreams: Vec<LivestreamModel> = sqlx::query_as("SELECT * FROM livestreams")
        .fetch_all(&mut *tx)
        .await?;

    // ランク算出
    let mut ranking = Vec::new();
    for livestream in livestreams {
        let MysqlDecimal(reactions) = sqlx::query_scalar("SELECT COUNT(*) FROM livestreams l INNER JOIN reactions r ON l.id = r.livestream_id WHERE l.id = ?")
            .bind(livestream.id)
            .fetch_one(&mut *tx)
            .await?;

        let MysqlDecimal(total_tips) = sqlx::query_scalar("SELECT IFNULL(SUM(l2.tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l.id = l2.livestream_id WHERE l.id = ?")
            .bind(livestream.id)
            .fetch_one(&mut *tx)
            .await?;

        let score = reactions + total_tips;
        ranking.push(LivestreamRankingEntry {
            livestream_id: livestream.id,
            score,
        })
    }
    ranking.sort_by(|a, b| {
        a.score
            .cmp(&b.score)
            .then_with(|| a.livestream_id.cmp(&b.livestream_id))
    });

    let rpos = ranking
        .iter()
        .rposition(|entry| entry.livestream_id == livestream_id)
        .unwrap();
    let rank = (ranking.len() - rpos) as i64;

    // 視聴者数算出
    let MysqlDecimal(viewers_count) = sqlx::query_scalar("SELECT COUNT(*) FROM livestreams l INNER JOIN livestream_viewers_history h ON h.livestream_id = l.id WHERE l.id = ?")
        .bind(livestream_id)
        .fetch_one(&mut *tx)
        .await?;

    // 最大チップ額
    let MysqlDecimal(max_tip) = sqlx::query_scalar("SELECT IFNULL(MAX(tip), 0) FROM livestreams l INNER JOIN livecomments l2 ON l2.livestream_id = l.id WHERE l.id = ?")
        .bind(livestream_id)
        .fetch_one(&mut *tx)
        .await?;

    // リアクション数
    let MysqlDecimal(total_reactions) = sqlx::query_scalar("SELECT COUNT(*) FROM livestreams l INNER JOIN reactions r ON r.livestream_id = l.id WHERE l.id = ?")
        .bind(livestream_id)
        .fetch_one(&mut *tx)
        .await?;

    // スパム報告数
    let MysqlDecimal(total_reports) = sqlx::query_scalar("SELECT COUNT(*) FROM livestreams l INNER JOIN livecomment_reports r ON r.livestream_id = l.id WHERE l.id = ?")
        .bind(livestream_id)
        .fetch_one(&mut *tx)
        .await?;

    tx.commit().await?;

    Ok(axum::Json(LivestreamStatistics {
        rank,
        viewers_count,
        max_tip,
        total_reactions,
        total_reports,
    }))
}

#[derive(Debug, serde::Serialize)]
struct PaymentResult {
    total_tip: i64,
}

async fn get_payment_result(
    State(AppState { pool, .. }): State<AppState>,
) -> Result<axum::Json<PaymentResult>, Error> {
    let mut tx = pool.begin().await?;

    let MysqlDecimal(total_tip) =
        sqlx::query_scalar("SELECT IFNULL(SUM(tip), 0) FROM livecomments")
            .fetch_one(&mut *tx)
            .await?;

    tx.commit().await?;

    Ok(axum::Json(PaymentResult { total_tip }))
}
