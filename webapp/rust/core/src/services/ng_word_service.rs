use crate::db::HaveDBPool;
use crate::models::livestream::LivestreamId;
use crate::models::ng_word::{CreateNgWord, NgWord, NgWordId};
use crate::models::user::UserId;
use crate::repos::livestream_comment_repository::{
    HaveLivestreamCommentRepository, LivestreamCommentRepository,
};
use crate::repos::ng_word_repository::{HaveNgWordRepository, NgWordRepository};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait NgWordService {
    async fn create(&self, ng_word: &CreateNgWord) -> ServiceResult<NgWordId>;
    async fn find_all_by_livestream_id_and_user_id(
        &self,
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> ServiceResult<Vec<NgWord>>;
}

pub trait HaveNgWordService {
    type Service: NgWordService;

    fn ng_word_service(&self) -> &Self::Service;
}

pub trait NgWordServiceImpl:
    Sync + HaveDBPool + HaveNgWordRepository + HaveLivestreamCommentRepository
{
}

#[async_trait]
impl<T: NgWordServiceImpl> NgWordService for T {
    async fn create(&self, ng_word: &CreateNgWord) -> ServiceResult<NgWordId> {
        let mut tx = self.get_db_pool().begin().await?;

        let word_id = self.ng_word_repo().create(&mut tx, ng_word).await?;

        let ng_words = self
            .ng_word_repo()
            .find_all_by_livestream_id(&mut tx, &ng_word.livestream_id)
            .await?;

        // NGワードにヒットする過去の投稿も全削除する
        let comment_repo = self.livestream_comment_repo();
        let comments = comment_repo.find_all(&mut tx).await?;
        for ngword in &ng_words {
            for comment in &comments {
                comment_repo
                    .remove_if_match_ng_word(&mut tx, comment, &ngword.word)
                    .await?;
            }
        }

        tx.commit().await?;

        Ok(word_id)
    }

    async fn find_all_by_livestream_id_and_user_id(
        &self,
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> ServiceResult<Vec<NgWord>> {
        let mut conn = self.get_db_pool().acquire().await?;
        let ng_words = self
            .ng_word_repo()
            .find_all_by_livestream_id_and_user_id_order_by_created_at(
                &mut conn,
                livestream_id,
                user_id,
            )
            .await?;

        Ok(ng_words)
    }
}
