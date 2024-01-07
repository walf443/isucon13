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

use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::user::{CreateUser, User, UserId};
use isupipe_core::repos::user_repository::UserRepository;

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

        let result = sqlx::query!(
            "INSERT INTO users (name, display_name, description, password) VALUES(?, ?, ?, ?)",
            &user.name,
            &user.display_name,
            &user.description,
            &hashed_password,
        )
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
        let user_model = sqlx::query_as!(User, "SELECT id, name, display_name, description, password as hashed_password FROM users WHERE id = ?", id)
            .fetch_optional(conn)
            .await?;

        Ok(user_model)
    }

    async fn find_all(&self, conn: &mut DBConn) -> isupipe_core::repos::Result<Vec<User>> {
        let users: Vec<User> = sqlx::query_as!(
            User,
            "SELECT id, name, display_name, description, password as hashed_password FROM users"
        )
        .fetch_all(conn)
        .await?;

        Ok(users)
    }

    async fn find_id_by_name(
        &self,
        conn: &mut DBConn,
        name: &str,
    ) -> isupipe_core::repos::Result<Option<UserId>> {
        let user_id: Option<UserId> =
            sqlx::query_scalar!("SELECT id as `id:UserId` FROM users WHERE name = ?", name)
                .fetch_optional(conn)
                .await?;

        Ok(user_id)
    }

    async fn find_by_name(
        &self,
        conn: &mut DBConn,
        name: &str,
    ) -> isupipe_core::repos::Result<Option<User>> {
        let user_model: Option<User> = sqlx::query_as!(User, "SELECT id, name, display_name, description, password as hashed_password FROM users WHERE name = ?", name)
            .fetch_optional(conn)
            .await?;

        Ok(user_model)
    }
}
