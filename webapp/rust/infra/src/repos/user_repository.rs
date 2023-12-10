use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::user::UserModel;
use isupipe_core::repos::user_repository::UserRepository;

pub struct UserRepositoryInfra {}

#[async_trait]
impl UserRepository for UserRepositoryInfra {
    async fn find_by_name(
        &self,
        conn: &mut DBConn,
        name: &str,
    ) -> isupipe_core::repos::Result<Option<UserModel>> {
        let user_model: Option<UserModel> = sqlx::query_as("SELECT * FROM users WHERE name = ?")
            .bind(name)
            .fetch_optional(conn)
            .await?;

        Ok(user_model)
    }
}
