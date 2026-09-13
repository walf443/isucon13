use crate::tables::livestream::LivestreamRow;
use async_trait::async_trait;
use isupipe_core::db::DBConn;
use isupipe_core::models::livestream::{CreateLivestream, Livestream, LivestreamId};
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_repository::LivestreamRepository;

#[derive(Clone)]
pub struct LivestreamRepositoryInfra {}

#[async_trait]
impl LivestreamRepository for LivestreamRepositoryInfra {
    async fn create<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        stream: &CreateLivestream,
    ) -> isupipe_core::repos::Result<LivestreamId> {
        let row = LivestreamRow::create()
            .user_id(&stream.user_id)
            .title(&stream.title)
            .description(&stream.description)
            .playlist_url(&stream.playlist_url)
            .thumbnail_url(&stream.thumbnail_url)
            .start_at(stream.start_at)
            .end_at(stream.end_at)
            .exec(conn)
            .await?;

        Ok(row.id)
    }

    async fn find_all<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
    ) -> isupipe_core::repos::Result<Vec<Livestream>> {
        let rows = LivestreamRow::all().exec(conn).await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }

    async fn find_all_order_by_id_desc<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
    ) -> isupipe_core::repos::Result<Vec<Livestream>> {
        let rows = LivestreamRow::all()
            .order_by(LivestreamRow::fields().id().desc())
            .exec(conn)
            .await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }

    async fn find_all_order_by_id_desc_limit<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        limit: i64,
    ) -> isupipe_core::repos::Result<Vec<Livestream>> {
        let rows = LivestreamRow::all()
            .order_by(LivestreamRow::fields().id().desc())
            .limit(limit as usize)
            .exec(conn)
            .await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }

    async fn find_all_by_user_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<Vec<Livestream>> {
        let rows = LivestreamRow::filter(LivestreamRow::fields().user_id().eq(user_id))
            .exec(conn)
            .await?;

        Ok(rows.into_iter().map(Into::into).collect())
    }

    async fn find<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        id: &LivestreamId,
    ) -> isupipe_core::repos::Result<Option<Livestream>> {
        let row = LivestreamRow::filter(LivestreamRow::fields().id().eq(id))
            .first()
            .exec(conn)
            .await?;

        Ok(row.map(Into::into))
    }

    async fn exist_by_id_and_user_id<'c>(
        &self,
        conn: &'c mut DBConn<'c>,
        id: &LivestreamId,
        user_id: &UserId,
    ) -> isupipe_core::repos::Result<bool> {
        let count = LivestreamRow::filter(LivestreamRow::fields().id().eq(id))
            .filter(LivestreamRow::fields().user_id().eq(user_id))
            .count()
            .exec(conn)
            .await?;

        Ok(count > 0)
    }
}
