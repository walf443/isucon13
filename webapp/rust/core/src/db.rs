use sqlx::{MySqlConnection, MySqlPool};
use sqlx::mysql::MySqlPoolOptions;

pub type DBPool = MySqlPool;
pub type DBConn = MySqlConnection;

pub trait HaveDBPool {
    fn get_db_pool(&self) -> &DBPool;
}
pub fn build_database_connection_options() -> sqlx::mysql::MySqlConnectOptions {
    _build_database_connection_options(false)
}

#[cfg(any(feature = "test", test))]
fn build_database_connection_options_for_test() -> sqlx::mysql::MySqlConnectOptions {
    _build_database_connection_options(true)
}

#[cfg(any(feature = "test", test))]
pub async fn get_db_pool() -> Result<DBPool,sqlx::Error> {
    let pool = MySqlPoolOptions::new().connect_with(build_database_connection_options_for_test()).await?;

    Ok(pool)
}

fn _build_database_connection_options(is_test_mode: bool) -> sqlx::mysql::MySqlConnectOptions {
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
    if is_test_mode {
        options = options.database("isupipe-test")
    }
    options
}
