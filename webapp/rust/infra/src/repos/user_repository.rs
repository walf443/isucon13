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

use crate::qbey_support::bind_qbey_values;
use crate::tables::user::TABLE_USERS;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::user::{CreateUser, User, UserId};
use isupipe_core::repos::user_repository::UserRepository;
use qbey_mysql::qbey;

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
        let t = &TABLE_USERS;
        let mut q = qbey(t.table());
        q.select(&TABLE_USERS.default_cols());
        q.and_where(t.id().eq(*id.inner()));
        let (sql, binds) = q.to_sql();
        let user_model = bind_qbey_values!(sqlx::query_as::<_, User>(&sql), binds)
            .fetch_optional(conn)
            .await?;

        Ok(user_model)
    }

    async fn find_all(&self, conn: &mut DBConn) -> isupipe_core::repos::Result<Vec<User>> {
        let t = &TABLE_USERS;
        let mut q = qbey(t.table());
        q.select(&TABLE_USERS.default_cols());
        let (sql, binds) = q.to_sql();
        let users = bind_qbey_values!(sqlx::query_as::<_, User>(&sql), binds)
            .fetch_all(conn)
            .await?;

        Ok(users)
    }

    async fn find_id_by_name(
        &self,
        conn: &mut DBConn,
        name: &str,
    ) -> isupipe_core::repos::Result<Option<UserId>> {
        let t = &TABLE_USERS;
        let mut q = qbey(t.table());
        q.select(&[t.id()]);
        q.and_where(t.name().eq(name));
        let (sql, binds) = q.to_sql();
        let user_id = bind_qbey_values!(sqlx::query_scalar::<_, UserId>(&sql), binds)
            .fetch_optional(conn)
            .await?;

        Ok(user_id)
    }

    async fn find_by_name(
        &self,
        conn: &mut DBConn,
        name: &str,
    ) -> isupipe_core::repos::Result<Option<User>> {
        let t = &TABLE_USERS;
        let mut q = qbey(t.table());
        q.select(&TABLE_USERS.default_cols());
        q.and_where(t.name().eq(name));
        let (sql, binds) = q.to_sql();
        let user_model = bind_qbey_values!(sqlx::query_as::<_, User>(&sql), binds)
            .fetch_optional(conn)
            .await?;

        Ok(user_model)
    }
}
