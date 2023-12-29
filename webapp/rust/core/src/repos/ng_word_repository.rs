use crate::db::DBConn;
use crate::models::livestream::LivestreamId;
use crate::models::ng_word::{CreateNgWord, NgWord, NgWordId};
use crate::models::user::UserId;
use crate::repos::Result;
use async_trait::async_trait;

#[async_trait]
pub trait NgWordRepository {
    async fn create(&self, conn: &mut DBConn, ng_word: &CreateNgWord) -> Result<NgWordId>;
    async fn find_all_by_livestream_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
    ) -> Result<Vec<NgWord>>;

    async fn find_all_by_livestream_id_and_user_id(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> Result<Vec<NgWord>>;

    async fn find_all_by_livestream_id_and_user_id_order_by_created_at(
        &self,
        conn: &mut DBConn,
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> Result<Vec<NgWord>>;
}

pub trait HaveNgWordRepository {
    type Repo: NgWordRepository;
    fn ng_word_repo(&self) -> &Self::Repo;
}
