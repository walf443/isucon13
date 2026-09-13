/// アプリケーション全体で共有する toasty のデータベースハンドル (コネクションプール)。
pub type DBPool = toasty::Db;

/// リポジトリが受け取るクエリ実行ハンドル。
///
/// `toasty::Db` / `toasty::Connection` / `toasty::Transaction` のいずれも渡せる。
/// `mockall` + `async_trait` と組み合わせるためにライフタイムを明示している。
/// リポジトリのメソッドでは `async fn f<'c>(&self, conn: &'c mut DBConn<'c>, ...)`
/// の形で受け取ること。
pub type DBConn<'a> = dyn toasty::Executor + 'a;

pub trait HaveDBPool {
    fn get_db_pool(&self) -> &DBPool;
}

/// 環境変数から MySQL の接続 URL を組み立てる。
pub fn build_database_url() -> String {
    let mut host = "127.0.0.1".to_string();
    let mut port = "3306".to_string();
    let mut user = "isucon".to_string();
    let mut password = "isucon".to_string();
    let mut database = "isupipe".to_string();

    if let Ok(v) = std::env::var("ISUCON13_MYSQL_DIALCONFIG_ADDRESS") {
        host = v;
    }
    if let Ok(v) = std::env::var("ISUCON13_MYSQL_DIALCONFIG_PORT") {
        port = v;
    }
    if let Ok(v) = std::env::var("ISUCON13_MYSQL_DIALCONFIG_USER") {
        user = v;
    }
    if let Ok(v) = std::env::var("ISUCON13_MYSQL_DIALCONFIG_PASSWORD") {
        password = v;
    }
    if let Ok(v) = std::env::var("ISUCON13_MYSQL_DIALCONFIG_DATABASE") {
        database = v;
    }

    format!(
        "mysql://{}:{}@{host}:{port}/{database}?collation=utf8mb4_general_ci",
        url_encode(&user),
        url_encode(&password)
    )
}

/// URL の userinfo 部分に含められない文字をパーセントエンコードする。
fn url_encode(s: &str) -> String {
    let mut out = String::with_capacity(s.len());
    for b in s.bytes() {
        match b {
            b'A'..=b'Z' | b'a'..=b'z' | b'0'..=b'9' | b'-' | b'_' | b'.' | b'~' => {
                out.push(b as char)
            }
            _ => out.push_str(&format!("%{b:02X}")),
        }
    }
    out
}

/// テスト用の DB ハンドルを返す。モデルは登録しないので、raw SQL とトランザクション制御にのみ使える。
/// infra 側のテストではモデルを登録した `Db` を別途組み立てること。
#[cfg(any(feature = "test", test))]
pub async fn get_db_pool() -> toasty::Result<DBPool> {
    let url = get_test_db_url().await;
    toasty::Db::builder().max_pool_size(2).connect(&url).await
}

#[cfg(any(feature = "test", test))]
static TEST_CONTAINER: tokio::sync::OnceCell<TestContainer> = tokio::sync::OnceCell::const_new();

#[cfg(any(feature = "test", test))]
#[dtor::dtor(unsafe)]
fn cleanup_containers() {
    if let Some(tc) = TEST_CONTAINER.get() {
        if let Ok(mut guard) = tc.container.lock() {
            if let Some(container) = guard.take() {
                if let Ok(rt) = tokio::runtime::Runtime::new() {
                    let _ = rt.block_on(container.rm());
                }
            }
        }
    }
}

/// テストで起動する MySQL コンテナのイメージタグ。
#[cfg(any(feature = "test", test))]
const MYSQL_IMAGE_TAG: &str = "8.1";

#[cfg(any(feature = "test", test))]
struct TestContainer {
    url: String,
    container:
        std::sync::Mutex<Option<testcontainers::ContainerAsync<testcontainers::GenericImage>>>,
}

// Safety: TestContainer is only accessed through OnceCell which provides synchronization
#[cfg(any(feature = "test", test))]
unsafe impl Send for TestContainer {}
#[cfg(any(feature = "test", test))]
unsafe impl Sync for TestContainer {}

/// テスト用 MySQL コンテナを (未起動なら) 起動し、スキーマ適用済みの接続 URL を返す。
#[cfg(any(feature = "test", test))]
pub async fn get_test_db_url() -> String {
    let tc = TEST_CONTAINER
        .get_or_init(|| async {
            use testcontainers::core::{IntoContainerPort as _, WaitFor};
            use testcontainers::runners::AsyncRunner as _;
            use testcontainers::{GenericImage, ImageExt as _};

            let container = GenericImage::new("mysql", MYSQL_IMAGE_TAG)
                .with_exposed_port(3306.tcp())
                .with_wait_for(WaitFor::message_on_stderr(
                    "X Plugin ready for connections. Bind-address",
                ))
                .with_wait_for(WaitFor::message_on_stderr(
                    "/usr/sbin/mysqld: ready for connections.",
                ))
                .with_env_var("MYSQL_DATABASE", "test")
                .with_env_var("MYSQL_ALLOW_EMPTY_PASSWORD", "yes")
                .start()
                .await
                .unwrap();
            let host_port = container.get_host_port_ipv4(3306).await.unwrap();

            let url = format!(
                "mysql://root@127.0.0.1:{}/test?collation=utf8mb4_general_ci",
                host_port
            );

            // Create schema using a temporary handle
            let mut db = toasty::Db::builder()
                .max_pool_size(1)
                .connect(&url)
                .await
                .unwrap();
            init_schema(&mut db).await;

            TestContainer {
                url,
                container: std::sync::Mutex::new(Some(container)),
            }
        })
        .await;

    tc.url.clone()
}

#[cfg(any(feature = "test", test))]
async fn init_schema(db: &mut DBPool) {
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
        toasty::sql::statement(trimmed).exec(db).await.unwrap();
    }
}
