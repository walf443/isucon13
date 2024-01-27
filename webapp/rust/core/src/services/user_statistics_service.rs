use crate::db::HaveDBPool;
use crate::models::livestream::Livestream;
use crate::models::user::{User, UserName};
use crate::models::user_ranking_entry::UserRankingEntry;
use crate::models::user_statistics::UserStatistics;
use crate::repos::livestream_comment_repository::{
    HaveLivestreamCommentRepository, LivestreamCommentRepository,
};
use crate::repos::livestream_repository::{HaveLivestreamRepository, LivestreamRepository};
use crate::repos::livestream_viewers_history_repository::{
    HaveLivestreamViewersHistoryRepository, LivestreamViewersHistoryRepository,
};
use crate::repos::reaction_repository::{HaveReactionRepository, ReactionRepository};
use crate::repos::user_repository::{HaveUserRepository, UserRepository};
use crate::services::ServiceResult;
use async_trait::async_trait;

#[async_trait]
pub trait UserStatisticsService {
    async fn get_stats(&self, user: &User) -> ServiceResult<UserStatistics>;
    async fn get_rank(&self, user_name: &UserName) -> ServiceResult<i64>;
    async fn get_viewers_count(&self, user_livestreams: &[Livestream]) -> ServiceResult<i64>;
    async fn get_total_reactions(&self, user_name: &UserName) -> ServiceResult<i64>;
    async fn get_total_comment_stats(
        &self,
        user_livestreams: &[Livestream],
    ) -> ServiceResult<(i64, i64)>;
    async fn get_favorite_emoji(&self, user_name: &UserName) -> ServiceResult<String>;
}

pub trait HaveUserStatisticsService {
    type Service: UserStatisticsService;

    fn user_statistics_service(&self) -> &Self::Service;
}

pub trait UserStatisticsServiceImpl:
    Sync
    + HaveDBPool
    + HaveReactionRepository
    + HaveLivestreamCommentRepository
    + HaveLivestreamRepository
    + HaveLivestreamViewersHistoryRepository
    + HaveUserRepository
{
}

#[async_trait]
impl<T: UserStatisticsServiceImpl> UserStatisticsService for T {
    async fn get_stats(&self, user: &User) -> ServiceResult<UserStatistics> {
        let mut tx = self.get_db_pool().begin().await?;

        let livestreams = self
            .livestream_repo()
            .find_all_by_user_id(&mut tx, &user.id)
            .await?;

        let (total_live_comments, total_tip) = self.get_total_comment_stats(&livestreams).await?;

        let user_stats = UserStatistics {
            rank: self.get_rank(&user.name).await?,
            viewers_count: self.get_viewers_count(&livestreams).await?,
            total_reactions: self.get_total_reactions(&user.name).await?,
            total_livecomments: total_live_comments,
            total_tip,
            favorite_emoji: self.get_favorite_emoji(&user.name).await?,
        };

        Ok(user_stats)
    }

    async fn get_rank(&self, user_name: &UserName) -> ServiceResult<i64> {
        let mut tx = self.get_db_pool().begin().await?;

        // ランク算出
        let users = self.user_repo().find_all(&mut tx).await?;

        let mut ranking = Vec::new();
        for user in users {
            let reaction_count = self
                .reaction_repo()
                .count_by_livestream_user_id(&mut tx, &user.id)
                .await?;

            let tips = self
                .livestream_comment_repo()
                .get_sum_tip_of_livestream_user_id(&mut tx, &user.id)
                .await?;

            let score = reaction_count + tips;
            ranking.push(UserRankingEntry {
                username: user.name,
                score,
            });
        }
        ranking.sort_by(|a, b| {
            a.score
                .cmp(&b.score)
                .then_with(|| a.username.inner().cmp(&b.username.inner()))
        });

        let rpos = ranking
            .iter()
            .rposition(|entry| entry.username.inner() == user_name.inner())
            .unwrap();
        let rank = (ranking.len() - rpos) as i64;

        Ok(rank)
    }

    async fn get_viewers_count(&self, user_livestreams: &[Livestream]) -> ServiceResult<i64> {
        let mut tx = self.get_db_pool().begin().await?;

        // 合計視聴者数
        let mut viewers_count = 0;
        for livestream in user_livestreams {
            let cnt = self
                .livestream_viewers_history_repo()
                .count_by_livestream_id(&mut tx, &livestream.id)
                .await?;
            viewers_count += cnt;
        }

        Ok(viewers_count)
    }

    async fn get_total_reactions(&self, user_name: &UserName) -> ServiceResult<i64> {
        let mut conn = self.get_db_pool().acquire().await?;

        let total_reactions = self
            .reaction_repo()
            .count_by_livestream_user_name(&mut conn, user_name)
            .await?;
        Ok(total_reactions)
    }

    // ライブコメント数、チップ合計
    async fn get_total_comment_stats(
        &self,
        user_livestreams: &[Livestream],
    ) -> ServiceResult<(i64, i64)> {
        let mut tx = self.get_db_pool().begin().await?;

        let mut total_livecomments = 0;
        let mut total_tip = 0;

        for livestream in user_livestreams {
            let comments = self
                .livestream_comment_repo()
                .find_all_by_livestream_id(&mut tx, &livestream.id)
                .await?;

            for comment in comments {
                total_tip += comment.tip;
                total_livecomments += 1;
            }
        }

        Ok((total_livecomments, total_tip))
    }

    async fn get_favorite_emoji(&self, user_name: &UserName) -> ServiceResult<String> {
        let mut conn = self.get_db_pool().acquire().await?;

        let favorite_emoji = self
            .reaction_repo()
            .most_favorite_emoji_by_livestream_user_name(&mut conn, user_name)
            .await?;
        Ok(favorite_emoji)
    }
}
