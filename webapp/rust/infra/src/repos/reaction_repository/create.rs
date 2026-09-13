use crate::repos::reaction_repository::ReactionRepositoryInfra;
use crate::tables::reaction::ReactionRow;
use crate::test_support::get_db_pool;
use fake::{Fake, Faker};
use isupipe_core::models::reaction::{CreateReaction, Reaction};
use isupipe_core::repos::reaction_repository::ReactionRepository;

#[tokio::test]
async fn success_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let repo = ReactionRepositoryInfra {};

    let reaction: CreateReaction = Faker.fake();
    let reaction_id = repo.create(&mut tx, &reaction).await.unwrap();

    let got: Reaction = ReactionRow::all()
        .filter(ReactionRow::fields().id().eq(*reaction_id.inner()))
        .one()
        .exec(&mut tx)
        .await
        .unwrap()
        .into();

    assert_eq!(got.id, reaction_id);
    assert_eq!(got.user_id, reaction.user_id);
    assert_eq!(got.emoji_name, reaction.emoji_name);
    assert_eq!(got.livestream_id, reaction.livestream_id);
    assert_eq!(got.created_at, reaction.created_at);
}
