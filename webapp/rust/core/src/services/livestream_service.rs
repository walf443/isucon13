use crate::db::HaveDBPool;
use crate::models::livestream::{CreateLivestream, Livestream, LivestreamId};
use crate::models::tag::TagId;
use crate::models::user::UserId;
use crate::repos::livestream_repository::{HaveLivestreamRepository, LivestreamRepository};
use crate::repos::livestream_tag_repository::{
    HaveLivestreamTagRepository, LivestreamTagRepository,
};
use crate::repos::reservation_slot_repository::{
    HaveReservationSlotRepository, ReservationSlotRepository,
};
use crate::repos::tag_repository::{HaveTagRepository, TagRepository};
use crate::services::ServiceError::InvalidReservationRange;
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamService {
    async fn create(
        &self,
        livestream: &CreateLivestream,
        tag_ids: &[TagId],
    ) -> ServiceResult<Livestream>;
    async fn find(&self, livestream_id: &LivestreamId) -> ServiceResult<Option<Livestream>>;

    async fn find_recent_livestreams(&self, limit: Option<i64>) -> ServiceResult<Vec<Livestream>>;
    async fn find_recent_by_tag_name(&self, tag_name: &str) -> ServiceResult<Vec<Livestream>>;
    async fn find_all_by_user_id(&self, user_id: &UserId) -> ServiceResult<Vec<Livestream>>;

    async fn exist_by_id_and_user_id(
        &self,
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> ServiceResult<bool>;
}

pub trait HaveLivestreamService {
    type Service: LivestreamService;

    fn livestream_service(&self) -> &Self::Service;
}

pub trait LivestreamServiceImpl:
    Sync
    + HaveDBPool
    + HaveLivestreamRepository
    + HaveLivestreamTagRepository
    + HaveTagRepository
    + HaveReservationSlotRepository
{
}

#[async_trait]
impl<T: LivestreamServiceImpl> LivestreamService for T {
    async fn create(
        &self,
        livestream: &CreateLivestream,
        tag_ids: &[TagId],
    ) -> ServiceResult<Livestream> {
        let mut tx = self.get_db_pool().begin().await?;

        let reservation_slot_repo = self.reservation_slot_repo();

        let slots = reservation_slot_repo
            .find_all_between_for_update(&mut tx, livestream.start_at, livestream.end_at)
            .await
            .map_err(|e| {
                tracing::warn!("予約枠一覧取得でエラー発生: {e:?}");
                e
            })?;

        for slot in slots {
            let count = reservation_slot_repo
                .find_slot_between(&mut tx, slot.start_at, slot.end_at)
                .await?;
            tracing::info!(
                "{} ~ {}予約枠の残数 = {}",
                slot.start_at,
                slot.end_at,
                slot.slot
            );
            if count < 1 {
                return Err(InvalidReservationRange);
            }
        }

        reservation_slot_repo
            .decrement_slot_between(&mut tx, livestream.start_at, livestream.end_at)
            .await?;

        let livestream_id = self.livestream_repo().create(&mut tx, livestream).await?;

        let tag_repo = self.livestream_tag_repo();
        for tag_id in tag_ids {
            tag_repo.insert(&mut tx, &livestream_id, tag_id).await?;
        }

        tx.commit().await?;

        Ok(Livestream {
            id: livestream_id,
            user_id: livestream.user_id.clone(),
            title: livestream.title.clone(),
            description: livestream.description.clone(),
            playlist_url: livestream.playlist_url.clone(),
            thumbnail_url: livestream.thumbnail_url.clone(),
            start_at: livestream.start_at,
            end_at: livestream.end_at,
        })
    }

    async fn find(&self, livestream_id: &LivestreamId) -> ServiceResult<Option<Livestream>> {
        let mut conn = self.get_db_pool().acquire().await?;
        let result = self
            .livestream_repo()
            .find(&mut conn, livestream_id)
            .await?;
        Ok(result)
    }

    async fn find_recent_livestreams(&self, limit: Option<i64>) -> ServiceResult<Vec<Livestream>> {
        let mut conn = self.get_db_pool().acquire().await?;
        let livestreams = match limit {
            None => {
                self.livestream_repo()
                    .find_all_order_by_id_desc(&mut conn)
                    .await?
            }
            Some(limit) => {
                self.livestream_repo()
                    .find_all_order_by_id_desc_limit(&mut conn, limit)
                    .await?
            }
        };

        Ok(livestreams)
    }

    async fn find_recent_by_tag_name(&self, tag_name: &str) -> ServiceResult<Vec<Livestream>> {
        let mut tx = self.get_db_pool().acquire().await?;
        let tag_id_list = self.tag_repo().find_ids_by_name(&mut tx, &tag_name).await?;

        let key_tagged_livestreams = self
            .livestream_tag_repo()
            .find_all_by_tag_ids(&mut tx, &tag_id_list)
            .await?;

        let mut livestream_models = Vec::new();
        let livestream_repo = self.livestream_repo();
        for key_tagged_livestream in key_tagged_livestreams {
            let ls = livestream_repo
                .find(&mut tx, &key_tagged_livestream.livestream_id)
                .await?
                .unwrap();

            livestream_models.push(ls);
        }

        Ok(livestream_models)
    }

    async fn find_all_by_user_id(&self, user_id: &UserId) -> ServiceResult<Vec<Livestream>> {
        let mut conn = self.get_db_pool().acquire().await?;
        let livestreams = self
            .livestream_repo()
            .find_all_by_user_id(&mut conn, user_id)
            .await?;

        Ok(livestreams)
    }

    async fn exist_by_id_and_user_id(
        &self,
        livestream_id: &LivestreamId,
        user_id: &UserId,
    ) -> ServiceResult<bool> {
        let mut conn = self.get_db_pool().acquire().await?;
        let is_exist = self
            .livestream_repo()
            .exist_by_id_and_user_id(&mut conn, livestream_id, user_id)
            .await?;
        Ok(is_exist)
    }
}
