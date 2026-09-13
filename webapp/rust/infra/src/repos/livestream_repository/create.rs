use crate::repos::livestream_repository::LivestreamRepositoryInfra;
use crate::test_support::InsertUserSetup;
use crate::tables::livestream::LivestreamRow;
use fake::{Fake, Faker};
use crate::test_support::get_db_pool;
use isupipe_core::models::livestream::{CreateLivestream, Livestream};
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::livestream_repository::LivestreamRepository;

#[tokio::test]
async fn success_case() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
                InsertUserSetup { id: 1, user: &user }.insert(&mut tx).await;
    }

    let mut input: CreateLivestream = Faker.fake();
    input.user_id = UserId::new(1);

    let repo = LivestreamRepositoryInfra {};
    let livestream_id = repo.create(&mut tx, &input).await.unwrap();

    let got: Livestream = LivestreamRow::all()
        .filter(LivestreamRow::fields().id().eq(*livestream_id.inner()))
        .one()
        .exec(&mut tx)
        .await
        .unwrap()
        .into();

    assert_eq!(got.id, livestream_id);
    assert_eq!(got.user_id, input.user_id);
    assert_eq!(got.title, input.title);
    assert_eq!(got.description, input.description);
    assert_eq!(got.playlist_url, input.playlist_url);
    assert_eq!(got.thumbnail_url, input.thumbnail_url);
    assert_eq!(got.start_at, input.start_at);
    assert_eq!(got.end_at, input.end_at);
}
