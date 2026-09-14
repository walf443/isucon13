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

use crate::tables::user::UserRow;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::user::{CreateUser, User, UserId, UserName};
use isupipe_core::repos::user_repository::UserRepository;

#[derive(Clone)]
pub struct UserRepositoryInfra {}

#[async_trait]
impl UserRepository for UserRepositoryInfra {
    async fn create<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        user: &CreateUser,
    ) -> isupipe_core::repos::Result<UserId> {
        let hashed_password = self.hash_password(&user.password)?;

        let row = UserRow::create()
            .name(UserName::new(user.name.clone()))
            .display_name(&user.display_name)
            .description(&user.description)
            .hashed_password(hashed_password)
            .exec(conn)
            .await?;

        Ok(row.id)
    }

    async fn find<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        id: &UserId,
    ) -> isupipe_core::repos::Result<Option<User>> {
        let row = UserRow::filter(UserRow::fields().id().eq(id))
            .first()
            .exec(conn)
            .await?;

        Ok(row.map(Into::into))
    }

    async fn find_all<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
    ) -> isupipe_core::repos::Result<Vec<User>> {
        let rows = UserRow::all().exec(conn).await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }

    async fn find_id_by_name<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        name: &str,
    ) -> isupipe_core::repos::Result<Option<UserId>> {
        let ids = UserRow::filter(UserRow::fields().name().eq(UserName::new(name.to_owned())))
            .limit(1)
            .select(UserRow::fields().id())
            .exec(conn)
            .await?;

        Ok(ids.into_iter().next())
    }

    async fn find_by_name<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        name: &str,
    ) -> isupipe_core::repos::Result<Option<User>> {
        let row = UserRow::filter(UserRow::fields().name().eq(UserName::new(name.to_owned())))
            .first()
            .exec(conn)
            .await?;

        Ok(row.map(Into::into))
    }
}
