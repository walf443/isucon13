use crate::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use crate::tables::livestream_comment::LivestreamCommentRow;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::livestream_comment::{CreateLivestreamComment, LivestreamComment};
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;

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

    let mut input: CreateLivestreamComment = Faker.fake();
    input.user_id = UserId::new(1);
    input.livestream_id = LivestreamId::new(1);

    let repo = LivestreamCommentRepositoryInfra {};
    let comment_id = repo.create(&mut tx, &input).await.unwrap();

    let got: LivestreamComment = LivestreamCommentRow::all()
        .filter(LivestreamCommentRow::fields().id().eq(*comment_id.inner()))
        .one()
        .exec(&mut tx)
        .await
        .unwrap()
        .into();

    assert_eq!(got.id, comment_id);
    assert_eq!(got.user_id, input.user_id);
    assert_eq!(got.livestream_id, input.livestream_id);
    assert_eq!(got.comment, input.comment);
    assert_eq!(got.tip, input.tip);
    assert_eq!(got.created_at, input.created_at);
}
