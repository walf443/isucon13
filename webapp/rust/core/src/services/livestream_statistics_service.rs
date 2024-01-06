use crate::db::HaveDBPool;
use crate::models::livestream::{Livestream, LivestreamId};
use crate::models::livestream_ranking_entry::LivestreamRankingEntry;
use crate::models::livestream_statistics::LivestreamStatistics;
use crate::repos::livestream_comment_report_repository::{
    HaveLivestreamCommentReportRepository, LivestreamCommentReportRepository,
};
use crate::repos::livestream_comment_repository::{
    HaveLivestreamCommentRepository, LivestreamCommentRepository,
};
use crate::repos::livestream_repository::{HaveLivestreamRepository, LivestreamRepository};
use crate::repos::livestream_viewers_history_repository::{
    HaveLivestreamViewersHistoryRepository, LivestreamViewersHistoryRepository,
};
use crate::repos::reaction_repository::{HaveReactionRepository, ReactionRepository};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait LivestreamStatisticsService {
    async fn get_stats(&self, livestream: &Livestream) -> ServiceResult<LivestreamStatistics>;
    async fn get_rank(&self, livestream: &Livestream) -> ServiceResult<i64>;
    async fn get_viewers_count(&self, livestream: &Livestream) -> ServiceResult<i64>;
    async fn get_max_tip(&self, livestream: &Livestream) -> ServiceResult<i64>;
    async fn get_total_reactions(&self, livestream: &Livestream) -> ServiceResult<i64>;
    async fn get_total_report(&self, livestream: &Livestream) -> ServiceResult<i64>;
}

pub trait HaveLivestreamStatisticsService {
    type Service: LivestreamStatisticsService;

    fn livestream_statistics_service(&self) -> &Self::Service;
}

pub trait LivestreamStatisticsServiceImpl:
    Sync
    + HaveDBPool
    + HaveLivestreamCommentReportRepository
    + HaveReactionRepository
    + HaveLivestreamCommentRepository
    + HaveLivestreamViewersHistoryRepository
    + HaveLivestreamRepository
{
}

#[async_trait]
impl<T: LivestreamStatisticsServiceImpl> LivestreamStatisticsService for T {
    async fn get_stats(&self, livestream: &Livestream) -> ServiceResult<LivestreamStatistics> {
        let stat = LivestreamStatistics {
            rank: self.get_rank(livestream).await?,
            viewers_count: self.get_viewers_count(livestream).await?,
            total_reactions: self.get_total_reactions(livestream).await?,
            total_reports: self.get_total_report(livestream).await?,
            max_tip: self.get_max_tip(livestream).await?,
        };

        Ok(stat)
    }

    /// ランク算出
    async fn get_rank(&self, livestream: &Livestream) -> ServiceResult<i64> {
        let mut conn = self.get_db_pool().acquire().await?;

        let livestreams = self.livestream_repo().find_all(&mut conn).await?;

        let mut ranking = Vec::new();

        for livestream in livestreams {
            let reactions = self
                .reaction_repo()
                .count_by_livestream_id(&mut conn, &livestream.id)
                .await?;

            let total_tips = self
                .livestream_comment_repo()
                .get_sum_tip_of_livestream_id(&mut conn, &livestream.id)
                .await?;

            let score = reactions + total_tips;
            ranking.push(LivestreamRankingEntry {
                livestream_id: LivestreamId::new(livestream.id.get()),
                score,
            })
        }

        ranking.sort_by(|a, b| {
            a.score
                .cmp(&b.score)
                .then_with(|| a.livestream_id.get().cmp(&b.livestream_id.get()))
        });

        let rpos = ranking
            .iter()
            .rposition(|entry| entry.livestream_id.get() == livestream.id.get())
            .unwrap();
        let rank = (ranking.len() - rpos) as i64;

        Ok(rank)
    }

    /// 視聴者数算出
    async fn get_viewers_count(&self, livestream: &Livestream) -> ServiceResult<i64> {
        let mut conn = self.get_db_pool().acquire().await?;
        let viewers_count = self
            .livestream_viewers_history_repo()
            .count_by_livestream_id(&mut conn, &livestream.id)
            .await?;
        Ok(viewers_count)
    }

    /// 最大チップ額
    async fn get_max_tip(&self, livestream: &Livestream) -> ServiceResult<i64> {
        let mut conn = self.get_db_pool().acquire().await?;
        let max_tip = self
            .livestream_comment_repo()
            .get_max_tip_of_livestream_id(&mut conn, &livestream.id)
            .await?;
        Ok(max_tip)
    }

    /// リアクション数
    async fn get_total_reactions(&self, livestream: &Livestream) -> ServiceResult<i64> {
        let mut conn = self.get_db_pool().acquire().await?;
        let total_reactions = self
            .reaction_repo()
            .count_by_livestream_id(&mut conn, &livestream.id)
            .await?;
        Ok(total_reactions)
    }

    /// スパム報告数
    async fn get_total_report(&self, livestream: &Livestream) -> ServiceResult<i64> {
        let mut conn = self.get_db_pool().acquire().await?;
        let total_reports = self
            .livestream_comment_report_repo()
            .count_by_livestream_id(&mut conn, &livestream.id)
            .await?;

        Ok(total_reports)
    }
}
