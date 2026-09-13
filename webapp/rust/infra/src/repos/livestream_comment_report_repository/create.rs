use crate::repos::livestream_comment_report_repository::LivestreamCommentReportRepositoryInfra;
use crate::tables::livestream_comment_report::LivestreamCommentReportRow;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertCommentSetup, InsertLivestreamSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::livestream_comment::{CreateLivestreamComment, LivestreamCommentId};
use isupipe_core::models::livestream_comment_report::{
    CreateLivestreamCommentReport, LivestreamCommentReport,
};
use isupipe_core::models::user::{CreateUser, UserId};
use isupipe_core::repos::livestream_comment_report_repository::LivestreamCommentReportRepository;

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

    let comment: CreateLivestreamComment = Faker.fake();
    {
        InsertCommentSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            comment: &comment,
        }
        .insert(&mut tx)
        .await;
    }

    let mut input: CreateLivestreamCommentReport = Faker.fake();
    input.user_id = UserId::new(1);
    input.livestream_id = LivestreamId::new(1);
    input.livestream_comment_id = LivestreamCommentId::new(1);

    let repo = LivestreamCommentReportRepositoryInfra {};
    let report_id = repo.create(&mut tx, &input).await.unwrap();

    let got: LivestreamCommentReport = LivestreamCommentReportRow::filter(
        LivestreamCommentReportRow::fields().id().eq(&report_id),
    )
    .one()
    .exec(&mut tx)
    .await
    .unwrap()
    .into();

    assert_eq!(got.id, report_id);
    assert_eq!(got.user_id, input.user_id);
    assert_eq!(got.livestream_id, input.livestream_id);
    assert_eq!(got.livestream_comment_id, input.livestream_comment_id);
    assert_eq!(got.created_at, input.created_at);
}
