use sqlx::{MySqlConnection, MySqlPool};

pub type DBPool = MySqlPool;
pub type DBConn = MySqlConnection;
