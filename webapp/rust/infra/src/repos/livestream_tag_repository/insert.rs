use crate::repos::livestream_tag_repository::LivestreamTagRepositoryInfra;
use crate::tables::livestream_tag::LivestreamTagRow;
use crate::test_support::get_db_pool;
use crate::test_support::{InsertLivestreamSetup, InsertTagSetup, InsertUserSetup};
use fake::{Fake, Faker};
use isupipe_core::models::livestream::{CreateLivestream, LivestreamId};
use isupipe_core::models::livestream_tag::LivestreamTag;
use isupipe_core::models::tag::{Tag, TagId};
use isupipe_core::models::user::CreateUser;
use isupipe_core::repos::livestream_tag_repository::LivestreamTagRepository;

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

    let tag: Tag = Faker.fake();
    {
        InsertTagSetup { id: 1, tag: &tag }.insert(&mut tx).await;
    }

    let repo = LivestreamTagRepositoryInfra {};
    repo.insert(&mut tx, &LivestreamId::new(1), &TagId::new(1))
        .await
        .unwrap();

    let got: LivestreamTag = LivestreamTagRow::all()
        .filter(
            LivestreamTagRow::fields()
                .livestream_id()
                .eq(LivestreamId::new(1)),
        )
        .filter(LivestreamTagRow::fields().tag_id().eq(TagId::new(1)))
        .one()
        .exec(&mut tx)
        .await
        .unwrap()
        .into();

    assert_eq!(*got.livestream_id.inner(), 1);
    assert_eq!(*got.tag_id.inner(), 1);
}
