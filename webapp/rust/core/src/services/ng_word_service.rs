use crate::db::HaveDBPool;
use crate::models::livestream::LivestreamId;
use crate::models::ng_word::NgWord;
use crate::models::user::UserId;
use crate::repos::ng_word_repository::{HaveNgWordRepository, NgWordRepository};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait NgWordService {
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

pub trait NgWordServiceImpl: Sync + HaveDBPool + HaveNgWordRepository {}

#[async_trait]
impl<T: NgWordServiceImpl> NgWordService for T {
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
