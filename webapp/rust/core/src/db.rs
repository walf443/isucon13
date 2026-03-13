#[cfg(any(feature = "test", test))]
use sqlx::mysql::MySqlPoolOptions;
use sqlx::{MySqlConnection, MySqlPool};

pub type DBPool = MySqlPool;
pub type DBConn = MySqlConnection;

pub trait HaveDBPool {
    fn get_db_pool(&self) -> &DBPool;
}
pub fn build_database_connection_options() -> sqlx::mysql::MySqlConnectOptions {
    _build_database_connection_options(false)
}

#[cfg(any(feature = "test", test))]
pub async fn get_db_pool() -> Result<DBPool, sqlx::Error> {
    let pool = get_test_pool().await;
    Ok(pool)
}

#[cfg(any(feature = "test", test))]
static TEST_POOL: tokio::sync::OnceCell<TestContainer> = tokio::sync::OnceCell::const_new();

#[cfg(any(feature = "test", test))]
struct TestContainer {
    pool: MySqlPool,
    // Keep container alive for the lifetime of the test suite
    _container: testcontainers::ContainerAsync<testcontainers_modules::mysql::Mysql>,
}

// Safety: TestContainer is only accessed through OnceCell which provides synchronization
#[cfg(any(feature = "test", test))]
unsafe impl Send for TestContainer {}
#[cfg(any(feature = "test", test))]
unsafe impl Sync for TestContainer {}

#[cfg(any(feature = "test", test))]
async fn get_test_pool() -> MySqlPool {
    let tc = TEST_POOL
        .get_or_init(|| async {
            use testcontainers::runners::AsyncRunner;
            use testcontainers_modules::mysql::Mysql;

            let container = Mysql::default().start().await.unwrap();
            let host_port = container.get_host_port_ipv4(3306).await.unwrap();

            let url = format!("mysql://root@127.0.0.1:{}/test", host_port);
            let pool = MySqlPoolOptions::new()
                .max_connections(30)
                .acquire_timeout(std::time::Duration::from_secs(120))
                .connect(&url)
                .await
                .unwrap();

            // Create schema
            init_schema(&pool).await;

            TestContainer {
                pool,
                _container: container,
            }
        })
        .await;

    tc.pool.clone()
}

#[cfg(any(feature = "test", test))]
async fn init_schema(pool: &MySqlPool) {
    let schema = include_str!("../../../sql/initdb.d/10_schema.sql");
    for statement in schema.split(';') {
        let trimmed = statement.trim();
        if trimmed.is_empty() {
            continue;
        }
        // Skip USE statement since we're already connected to the test db
        if trimmed.to_uppercase().starts_with("USE ") {
            continue;
        }
        sqlx::query(trimmed).execute(pool).await.unwrap();
    }
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
