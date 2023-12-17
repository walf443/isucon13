use isupipe_core::models::livestream::{Livestream, LivestreamModel};
use isupipe_core::models::livestream_comment::{LivestreamComment, LivestreamCommentModel};
use isupipe_core::models::livestream_comment_report::{
    LivestreamCommentReport, LivestreamCommentReportModel,
};
use isupipe_core::models::reaction::{Reaction, ReactionModel};
use isupipe_core::models::tag::Tag;
use isupipe_core::models::theme::Theme;
use isupipe_core::models::user::{User, UserModel};
use isupipe_core::repos::icon_repository::IconRepository;
use isupipe_core::repos::livestream_tag_repository::LivestreamTagRepository;
use isupipe_core::repos::tag_repository::TagRepository;
use isupipe_core::repos::theme_repository::ThemeRepository;
use isupipe_core::repos::user_repository::UserRepository;
use isupipe_core::utils::UtilResult;
use isupipe_http_core::FALLBACK_IMAGE;
use isupipe_infra::repos::icon_repository::IconRepositoryInfra;
use isupipe_infra::repos::livestream_tag_repository::LivestreamTagRepositoryInfra;
use isupipe_infra::repos::tag_repository::TagRepositoryInfra;
use isupipe_infra::repos::theme_repository::ThemeRepositoryInfra;
use isupipe_infra::repos::user_repository::UserRepositoryInfra;
use sqlx::MySqlConnection;

pub async fn fill_user_response(
    tx: &mut MySqlConnection,
    user_model: UserModel,
) -> UtilResult<User> {
    let theme_repo = ThemeRepositoryInfra {};
    let theme_model = theme_repo.find_by_user_id(&mut *tx, user_model.id).await?;

    let icon_repo = IconRepositoryInfra {};
    let image = icon_repo
        .find_image_by_user_id(&mut *tx, user_model.id)
        .await?;

    let image = if let Some(image) = image {
        image
    } else {
        tokio::fs::read(FALLBACK_IMAGE).await?
    };
    use sha2::digest::Digest as _;
    let icon_hash = sha2::Sha256::digest(image);

    Ok(User {
        id: user_model.id,
        name: user_model.name,
        display_name: user_model.display_name,
        description: user_model.description,
        theme: Theme {
            id: theme_model.id,
            dark_mode: theme_model.dark_mode,
        },
        icon_hash: format!("{:x}", icon_hash),
    })
}
pub async fn fill_livestream_response(
    tx: &mut MySqlConnection,
    livestream_model: LivestreamModel,
) -> UtilResult<Livestream> {
    let user_repo = UserRepositoryInfra {};
    let owner_model = user_repo
        .find(&mut *tx, livestream_model.user_id)
        .await?
        .unwrap();
    let owner = fill_user_response(tx, owner_model).await?;

    let livestream_tag_repo = LivestreamTagRepositoryInfra {};
    let livestream_tag_models = livestream_tag_repo
        .find_all_by_livestream_id(&mut *tx, livestream_model.id)
        .await?;

    let tag_repo = TagRepositoryInfra {};
    let mut tags = Vec::with_capacity(livestream_tag_models.len());
    for livestream_tag_model in livestream_tag_models {
        let tag_model = tag_repo.find(&mut *tx, livestream_tag_model.tag_id).await?;
        tags.push(Tag {
            id: tag_model.id,
            name: tag_model.name,
        });
    }

    Ok(Livestream {
        id: livestream_model.id,
        owner,
        title: livestream_model.title,
        tags,
        description: livestream_model.description,
        playlist_url: livestream_model.playlist_url,
        thumbnail_url: livestream_model.thumbnail_url,
        start_at: livestream_model.start_at,
        end_at: livestream_model.end_at,
    })
}
pub async fn fill_livecomment_response(
    tx: &mut MySqlConnection,
    livecomment_model: LivestreamCommentModel,
) -> UtilResult<LivestreamComment> {
    let comment_owner_model: UserModel = sqlx::query_as("SELECT * FROM users WHERE id = ?")
        .bind(livecomment_model.user_id)
        .fetch_one(&mut *tx)
        .await?;
    let comment_owner = fill_user_response(&mut *tx, comment_owner_model).await?;

    let livestream_model: LivestreamModel =
        sqlx::query_as("SELECT * FROM livestreams WHERE id = ?")
            .bind(livecomment_model.livestream_id)
            .fetch_one(&mut *tx)
            .await?;
    let livestream = fill_livestream_response(&mut *tx, livestream_model).await?;

    Ok(LivestreamComment {
        id: livecomment_model.id,
        user: comment_owner,
        livestream,
        comment: livecomment_model.comment,
        tip: livecomment_model.tip,
        created_at: livecomment_model.created_at,
    })
}
pub async fn fill_reaction_response(
    tx: &mut MySqlConnection,
    reaction_model: ReactionModel,
) -> UtilResult<Reaction> {
    let user_model: UserModel = sqlx::query_as("SELECT * FROM users WHERE id = ?")
        .bind(reaction_model.user_id)
        .fetch_one(&mut *tx)
        .await?;
    let user = fill_user_response(&mut *tx, user_model).await?;

    let livestream_model: LivestreamModel =
        sqlx::query_as("SELECT * FROM livestreams WHERE id = ?")
            .bind(reaction_model.livestream_id)
            .fetch_one(&mut *tx)
            .await?;
    let livestream = fill_livestream_response(&mut *tx, livestream_model).await?;

    Ok(Reaction {
        id: reaction_model.id,
        emoji_name: reaction_model.emoji_name,
        user,
        livestream,
        created_at: reaction_model.created_at,
    })
}
pub async fn fill_livecomment_report_response(
    tx: &mut MySqlConnection,
    report_model: LivestreamCommentReportModel,
) -> UtilResult<LivestreamCommentReport> {
    let reporter_model: UserModel = sqlx::query_as("SELECT * FROM users WHERE id = ?")
        .bind(report_model.user_id)
        .fetch_one(&mut *tx)
        .await?;
    let reporter = fill_user_response(&mut *tx, reporter_model).await?;

    let livecomment_model: LivestreamCommentModel =
        sqlx::query_as("SELECT * FROM livecomments WHERE id = ?")
            .bind(report_model.livecomment_id)
            .fetch_one(&mut *tx)
            .await?;
    let livecomment = fill_livecomment_response(&mut *tx, livecomment_model).await?;

    Ok(LivestreamCommentReport {
        id: report_model.id,
        reporter,
        livecomment,
        created_at: report_model.created_at,
    })
}
