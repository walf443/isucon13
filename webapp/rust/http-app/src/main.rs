use isupipe_http_core::routes::routes;
use isupipe_http_core::state::AppState;
use isupipe_infra::services::manager::ServiceManagerInfra;
use std::sync::Arc;

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

    let service = ServiceManagerInfra::new(pool.clone());

    let app = routes()
        .with_state(AppState {
            service,
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

#[derive(Debug, serde::Serialize)]
struct Session {
    id: String,
    user_id: i64,
    expires: i64,
}
