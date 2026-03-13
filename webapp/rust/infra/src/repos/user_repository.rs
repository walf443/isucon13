#[cfg(test)]
mod create;
#[cfg(test)]
mod find;
#[cfg(test)]
mod find_all;
#[cfg(test)]
mod find_by_name;
#[cfg(test)]
mod find_id_by_name;

use crate::sqipe_support::bind_sqipe_values;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::user::{CreateUser, User, UserId};
use isupipe_core::repos::user_repository::UserRepository;
use sqipe::{IntoColRef, col};
use sqipe_mysql::sqipe;

fn user_select_cols() -> Vec<sqipe::ColRef> {
    vec![
        col("id").into_col_ref(),
        col("name").into_col_ref(),
        col("display_name").into_col_ref(),
        col("description").into_col_ref(),
        col("password").as_("hashed_password"),
    ]
}

#[derive(Clone)]
pub struct UserRepositoryInfra {}

#[async_trait]
impl UserRepository for UserRepositoryInfra {
    async fn create(
        &self,
        conn: &mut DBConn,
        user: &CreateUser,
    ) -> isupipe_core::repos::Result<UserId> {
        let hashed_password = self.hash_password(&user.password)?;

        let result = sqlx::query(
            "INSERT INTO users (name, display_name, description, password) VALUES(?, ?, ?, ?)",
        )
        .bind(&user.name)
        .bind(&user.display_name)
        .bind(&user.description)
        .bind(&hashed_password)
        .execute(conn)
        .await?;

        let user_id = result.last_insert_id() as i64;

        Ok(UserId::new(user_id))
    }

    async fn find(
        &self,
        conn: &mut DBConn,
        id: &UserId,
    ) -> isupipe_core::repos::Result<Option<User>> {
        let mut q = sqipe("users");
        q.select_cols(&user_select_cols());
        q.and_where(("id", *id.inner()));
        let (sql, binds) = q.to_sql();
        let user_model = bind_sqipe_values!(sqlx::query_as::<_, User>(&sql), binds)
            .fetch_optional(conn)
            .await?;

        Ok(user_model)
    }

    async fn find_all(&self, conn: &mut DBConn) -> isupipe_core::repos::Result<Vec<User>> {
        let mut q = sqipe("users");
        q.select_cols(&user_select_cols());
        let (sql, binds) = q.to_sql();
        let users = bind_sqipe_values!(sqlx::query_as::<_, User>(&sql), binds)
            .fetch_all(conn)
            .await?;

        Ok(users)
    }

    async fn find_id_by_name(
        &self,
        conn: &mut DBConn,
        name: &str,
    ) -> isupipe_core::repos::Result<Option<UserId>> {
        let mut q = sqipe("users");
        q.select(&["id"]);
        q.and_where(("name", name));
        let (sql, binds) = q.to_sql();
        let user_id = bind_sqipe_values!(sqlx::query_scalar::<_, UserId>(&sql), binds)
            .fetch_optional(conn)
            .await?;

        Ok(user_id)
    }

    async fn find_by_name(
        &self,
        conn: &mut DBConn,
        name: &str,
    ) -> isupipe_core::repos::Result<Option<User>> {
        let mut q = sqipe("users");
        q.select_cols(&user_select_cols());
        q.and_where(("name", name));
        let (sql, binds) = q.to_sql();
        let user_model = bind_sqipe_values!(sqlx::query_as::<_, User>(&sql), binds)
            .fetch_optional(conn)
            .await?;

        Ok(user_model)
    }
}
