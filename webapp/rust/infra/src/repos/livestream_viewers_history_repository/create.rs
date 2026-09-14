use crate::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepositoryInfra;
use crate::tables::livestream_viewers_history::LivestreamViewersHistoryRow;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::livestream_viewers_history::CreateLivestreamViewersHistory;
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;

#[tokio::test]
async fn success_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        InsertUserSetup { id: 1, user: &user }.insert(&mut tx).await;
    }

    let stream: CreateLivestream = Faker.fake();
    {
        InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream,
        }
        .insert(&mut tx)
        .await;
    }

    let mut input: CreateLivestreamViewersHistory = Faker.fake();
    input.user_id = UserId::new(1);
    input.livestream_id = LivestreamId::new(1);

    let repo = LivestreamViewersHistoryRepositoryInfra {};
    repo.create(&mut tx, &input).await.unwrap();

    let count = LivestreamViewersHistoryRow::filter(
        LivestreamViewersHistoryRow::fields()
            .user_id()
            .eq(&input.user_id),
    )
    .filter(
        LivestreamViewersHistoryRow::fields()
            .livestream_id()
            .eq(&input.livestream_id),
    )
    .count()
    .exec(&mut tx)
    .await
    .unwrap();

    assert_eq!(count, 1);
}
