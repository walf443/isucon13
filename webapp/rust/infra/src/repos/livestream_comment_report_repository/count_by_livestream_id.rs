use crate::repos::livestream_comment_report_repository::LivestreamCommentReportRepositoryInfra;
use crate::test_support::get_db_pool;
use crate::test_support::{
    InsertCommentSetup, InsertLivestreamSetup, InsertReportSetup, InsertUserSetup,
};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::livestream_comment::CreateLivestreamComment;
use isupipe_core::models::livestream_comment_report::CreateLivestreamCommentReport;
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_comment_report_repository::LivestreamCommentReportRepository;

#[tokio::test]
async fn zero_case() {
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

    let repo = LivestreamCommentReportRepositoryInfra {};
    let result = repo
        .count_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result, 0);
}

#[tokio::test]
async fn counts_correctly() {
    let mut db = get_db_pool().await;
    let mut tx = db.transaction().await.unwrap();

    let user: CreateUser = Faker.fake();
    {
        InsertUserSetup { id: 1, user: &user }.insert(&mut tx).await;
    }

    let stream1: CreateLivestream = Faker.fake();
    let stream2: CreateLivestream = Faker.fake();
    {
        InsertLivestreamSetup {
            id: 1,
            user_id: 1,
            stream: &stream1,
        }
        .insert(&mut tx)
        .await;
        InsertLivestreamSetup {
            id: 2,
            user_id: 1,
            stream: &stream2,
        }
        .insert(&mut tx)
        .await;
    }

    let comment1: CreateLivestreamComment = Faker.fake();
    let comment2: CreateLivestreamComment = Faker.fake();
    {
        InsertCommentSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            comment: &comment1,
        }
        .insert(&mut tx)
        .await;
        InsertCommentSetup {
            id: 2,
            user_id: 1,
            livestream_id: 1,
            comment: &comment2,
        }
        .insert(&mut tx)
        .await;
    }

    {
        let r1: CreateLivestreamCommentReport = Faker.fake();
        InsertReportSetup {
            id: 1,
            user_id: 1,
            livestream_id: 1,
            livecomment_id: 1,
            report: &r1,
        }
        .insert(&mut tx)
        .await;
        let r2: CreateLivestreamCommentReport = Faker.fake();
        InsertReportSetup {
            id: 2,
            user_id: 1,
            livestream_id: 1,
            livecomment_id: 2,
            report: &r2,
        }
        .insert(&mut tx)
        .await;
        let r3: CreateLivestreamCommentReport = Faker.fake();
        InsertReportSetup {
            id: 3,
            user_id: 1,
            livestream_id: 2,
            livecomment_id: 1,
            report: &r3,
        }
        .insert(&mut tx)
        .await;
    }

    let repo = LivestreamCommentReportRepositoryInfra {};
    let result = repo
        .count_by_livestream_id(&mut tx, &LivestreamId::new(1))
        .await
        .unwrap();
    assert_eq!(result, 2);

    let result = repo
        .count_by_livestream_id(&mut tx, &LivestreamId::new(2))
        .await
        .unwrap();
    assert_eq!(result, 1);
}
