use crate::db::HaveDBPool;
use crate::models::livestream::LivestreamId;
use crate::models::livestream_comment::{
    CreateLivestreamComment, LivestreamComment, LivestreamCommentId,
};
use crate::repos::livestream_comment_repository::{
    HaveLivestreamCommentRepository, LivestreamCommentRepository,
};
use crate::repos::ng_word_repository::{HaveNgWordRepository, NgWordRepository};
use crate::services::ServiceError::CommentMatchSpam;
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamCommentService {
    async fn find(
        &self,
        livestream_comment_id: &LivestreamCommentId,
    ) -> ServiceResult<Option<LivestreamComment>>;

    async fn find_all_by_livestream_id(
        &self,
        livestream_id: &LivestreamId,
        limit: Option<i64>,
    ) -> ServiceResult<Vec<LivestreamComment>>;
    async fn get_sum_tip(&self) -> ServiceResult<i64>;

    async fn create(&self, comment: &CreateLivestreamComment) -> ServiceResult<LivestreamComment>;
}

pub trait HaveLivestreamCommentService {
    type Service: LivestreamCommentService;
    fn livestream_comment_service(&self) -> &Self::Service;
}

pub trait LivestreamCommentServiceImpl:
    Sync + HaveDBPool + HaveLivestreamCommentRepository + HaveNgWordRepository
{
}

#[async_trait]
impl<T: LivestreamCommentServiceImpl> LivestreamCommentService for T {
    async fn find(
        &self,
        livestream_comment_id: &LivestreamCommentId,
    ) -> ServiceResult<Option<LivestreamComment>> {
        let mut conn = self.get_db_pool().acquire().await?;
        let comment = self
            .livestream_comment_repo()
            .find(&mut conn, livestream_comment_id)
            .await?;
        Ok(comment)
    }

    async fn find_all_by_livestream_id(
        &self,
        livestream_id: &LivestreamId,
        limit: Option<i64>,
    ) -> ServiceResult<Vec<LivestreamComment>> {
        let mut conn = self.get_db_pool().acquire().await?;

        let comments = match limit {
            None => {
                self.livestream_comment_repo()
                    .find_all_by_livestream_id_order_by_created_at(&mut conn, livestream_id)
                    .await?
            }
            Some(limit) => {
                self.livestream_comment_repo()
                    .find_all_by_livestream_id_order_by_created_at_limit(
                        &mut conn,
                        livestream_id,
                        limit,
                    )
                    .await?
            }
        };

        Ok(comments)
    }

    async fn get_sum_tip(&self) -> ServiceResult<i64> {
        let mut conn = self.get_db_pool().acquire().await?;
        let sum_tip = self
            .livestream_comment_repo()
            .get_sum_tip(&mut conn)
            .await?;
        Ok(sum_tip)
    }

    async fn create(&self, comment: &CreateLivestreamComment) -> ServiceResult<LivestreamComment> {
        let mut tx = self.get_db_pool().begin().await?;

        let ng_word_repo = self.ng_word_repo();
        let ng_words = ng_word_repo
            .find_all_by_livestream_id_and_user_id(
                &mut tx,
                &comment.livestream_id,
                &comment.user_id,
            )
            .await?;

        for ngword in &ng_words {
            let hit_spam = ng_word_repo
                .count_by_ng_word_in_comment(&mut tx, &ngword.word, &comment.comment)
                .await?;
            tracing::info!("[hit_spam={}] comment = {}", hit_spam, &comment.comment);
            if hit_spam >= 1 {
                return Err(CommentMatchSpam);
            }
        }

        let comment_id = self
            .livestream_comment_repo()
            .create(&mut tx, comment)
            .await?;

        tx.commit().await?;

        Ok(LivestreamComment {
            id: comment_id,
            user_id: comment.user_id.clone(),
            livestream_id: comment.livestream_id.clone(),
            comment: comment.comment.clone(),
            tip: comment.tip,
            created_at: comment.created_at,
        })
    }
}
