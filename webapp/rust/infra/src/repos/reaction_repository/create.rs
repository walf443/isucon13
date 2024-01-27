use crate::repos::reaction_repository::ReactionRepositoryInfra;
use fake::{Fake, Faker};
use isupipe_core::db::get_db_pool;
use isupipe_core::models::reaction::{CreateReaction, Reaction};
use isupipe_core::repos::reaction_repository::ReactionRepository;

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();
    let mut tx = db_pool.begin().await.unwrap();

    let repo = ReactionRepositoryInfra {};

    let reaction: CreateReaction = Faker.fake();
    let reaction_id = repo.create(&mut tx, &reaction).await.unwrap();

    let got: Reaction = sqlx::query_as("SELECT * FROM reactions WHERE id = ?")
        .bind(&reaction_id)
        .fetch_one(&mut *tx)
        .await
        .unwrap();

    assert_eq!(got.id, reaction_id);
    assert_eq!(got.user_id, reaction.user_id);
    assert_eq!(got.emoji_name, reaction.emoji_name);
    assert_eq!(got.livestream_id, reaction.livestream_id);
    assert_eq!(got.created_at, reaction.created_at);
}
