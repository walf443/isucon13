use isupipe_core::models::livestream::{Livestream, LivestreamModel};
use isupipe_core::models::livestream_tag::LivestreamTagModel;
use isupipe_core::models::tag::{Tag, TagModel};
use isupipe_core::models::theme::{Theme, ThemeModel};
use isupipe_core::models::user::{User, UserModel};
use isupipe_http_core::FALLBACK_IMAGE;
use sqlx::MySqlConnection;

pub async fn fill_user_response(
    tx: &mut MySqlConnection,
    user_model: UserModel,
) -> sqlx::Result<User> {
    let theme_model: ThemeModel = sqlx::query_as("SELECT * FROM themes WHERE user_id = ?")
        .bind(user_model.id)
        .fetch_one(&mut *tx)
        .await?;

    let image: Option<Vec<u8>> = sqlx::query_scalar("SELECT image FROM icons WHERE user_id = ?")
        .bind(user_model.id)
        .fetch_optional(&mut *tx)
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
) -> sqlx::Result<Livestream> {
    let owner_model: UserModel = sqlx::query_as("SELECT * FROM users WHERE id = ?")
        .bind(livestream_model.user_id)
        .fetch_one(&mut *tx)
        .await?;
    let owner = fill_user_response(tx, owner_model).await?;

    let livestream_tag_models: Vec<LivestreamTagModel> =
        sqlx::query_as("SELECT * FROM livestream_tags WHERE livestream_id = ?")
            .bind(livestream_model.id)
            .fetch_all(&mut *tx)
            .await?;

    let mut tags = Vec::with_capacity(livestream_tag_models.len());
    for livestream_tag_model in livestream_tag_models {
        let tag_model: TagModel = sqlx::query_as("SELECT * FROM tags WHERE id = ?")
            .bind(livestream_tag_model.tag_id)
            .fetch_one(&mut *tx)
            .await?;
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