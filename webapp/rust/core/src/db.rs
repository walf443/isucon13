use sqlx::{MySqlConnection, MySqlPool};

pub type DBPool = MySqlPool;
pub type DBConn = MySqlConnection;

pub trait HaveDBPool {
    fn get_db_pool(&self) -> &DBPool;
}
