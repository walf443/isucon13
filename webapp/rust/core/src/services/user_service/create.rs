use crate::commands::{CommandError, CommandOutput};
use crate::db::get_db_pool;
use crate::models::user::{CreateUser, UserId};
use crate::repos::manager::tests::MockRepositoryManager;
use crate::repos::ReposError::TestError;
use crate::services::user_service::UserService;
use fake::{Fake, Faker};

#[tokio::test]
async fn user_repo_create_fail() {
    let db_pool = get_db_pool().await.unwrap();

    let mut service = MockRepositoryManager::new(db_pool);
    let user: CreateUser = Faker.fake();
    let dark_mode: bool = Faker.fake();
    let domain: String = Faker.fake();

    let got_user = user.clone();
    service
        .mock_user_repo
        .expect_create()
        .withf(move |_, u| u == &got_user)
        .returning(move |_, _| Err(TestError));

    let result = service.create(&user, dark_mode, &domain).await;
    assert!(result.is_err())
}

#[tokio::test]
async fn theme_repo_create_fail() {
    let db_pool = get_db_pool().await.unwrap();

    let mut service = MockRepositoryManager::new(db_pool);
    let user: CreateUser = Faker.fake();
    let dark_mode: bool = Faker.fake();
    let domain: String = Faker.fake();

    let expect_user_id: UserId = Faker.fake();
    let got_user = user.clone();
    let uid = expect_user_id.clone();
    service
        .mock_user_repo
        .expect_create()
        .withf(move |_, u| u == &got_user)
        .returning(move |_, _| Ok(uid.clone()));

    let uid = expect_user_id.clone();
    service
        .mock_theme_repo
        .expect_create()
        .withf(move |_, u, m| u.get() == uid.get() && m == &dark_mode)
        .returning(move |_, _, _| Err(TestError));

    let result = service.create(&user, dark_mode, &domain).await;
    assert!(result.is_err())
}

#[tokio::test]
async fn pdnsutil_command_add_record_error() {
    let db_pool = get_db_pool().await.unwrap();

    let mut service = MockRepositoryManager::new(db_pool);
    let user: CreateUser = Faker.fake();
    let dark_mode: bool = Faker.fake();
    let domain: String = Faker.fake();

    let expect_user_id: UserId = Faker.fake();
    let got_user = user.clone();
    let uid = expect_user_id.clone();
    service
        .mock_user_repo
        .expect_create()
        .withf(move |_, u| u == &got_user)
        .returning(move |_, _| Ok(uid.clone()));

    let uid = expect_user_id.clone();
    service
        .mock_theme_repo
        .expect_create()
        .withf(move |_, u, m| u.get() == uid.get() && m == &dark_mode)
        .returning(move |_, _, _| Ok(()));

    let got_user = user.clone();
    let dm = domain.clone();
    service
        .mock_pdns_util_command
        .expect_add_record()
        .withf(move |name, d| name == got_user.name && d == dm)
        .returning(move |_, _| Err(CommandError::TestError));

    let result = service.create(&user, dark_mode, &domain).await;
    assert!(result.is_err())
}

#[tokio::test]
async fn pdnsutil_command_add_record_fail() {
    let db_pool = get_db_pool().await.unwrap();

    let mut service = MockRepositoryManager::new(db_pool);
    let user: CreateUser = Faker.fake();
    let dark_mode: bool = Faker.fake();
    let domain: String = Faker.fake();

    let expect_user_id: UserId = Faker.fake();
    let got_user = user.clone();
    let uid = expect_user_id.clone();
    service
        .mock_user_repo
        .expect_create()
        .withf(move |_, u| u == &got_user)
        .returning(move |_, _| Ok(uid.clone()));

    let uid = expect_user_id.clone();
    service
        .mock_theme_repo
        .expect_create()
        .withf(move |_, u, m| u.get() == uid.get() && m == &dark_mode)
        .returning(move |_, _, _| Ok(()));

    let got_user = user.clone();
    let dm = domain.clone();
    service
        .mock_pdns_util_command
        .expect_add_record()
        .withf(move |name, d| name == got_user.name && d == dm)
        .returning(move |_, _| {
            Ok(CommandOutput {
                success: false,
                stdout: "stdout".as_bytes().to_vec(),
                stderr: "stderr".as_bytes().to_vec(),
            })
        });

    service
        .mock_user_repo
        .expect_hash_password()
        .returning(|_| Ok(Faker.fake()));

    let (u, out) = service.create(&user, dark_mode, &domain).await.unwrap();
    assert_eq!(out.success, false);
    assert_eq!(u.id.get(), expect_user_id.get())
}

#[tokio::test]
async fn success_case() {
    let db_pool = get_db_pool().await.unwrap();

    let mut service = MockRepositoryManager::new(db_pool);
    let user: CreateUser = Faker.fake();
    let dark_mode: bool = Faker.fake();
    let domain: String = Faker.fake();

    let expect_user_id: UserId = Faker.fake();
    let got_user = user.clone();
    let uid = expect_user_id.clone();
    service
        .mock_user_repo
        .expect_create()
        .withf(move |_, u| u == &got_user)
        .returning(move |_, _| Ok(uid.clone()));

    let uid = expect_user_id.clone();
    service
        .mock_theme_repo
        .expect_create()
        .withf(move |_, u, m| u.get() == uid.get() && m == &dark_mode)
        .returning(move |_, _, _| Ok(()));

    let got_user = user.clone();
    let dm = domain.clone();
    service
        .mock_pdns_util_command
        .expect_add_record()
        .withf(move |name, d| name == got_user.name && d == dm)
        .returning(move |_, _| {
            Ok(CommandOutput {
                success: true,
                stdout: "stdout".as_bytes().to_vec(),
                stderr: "stderr".as_bytes().to_vec(),
            })
        });

    service
        .mock_user_repo
        .expect_hash_password()
        .returning(|_| Ok(Faker.fake()));

    let (u, out) = service.create(&user, dark_mode, &domain).await.unwrap();
    assert_eq!(out.success, true);
    assert_eq!(u.id.get(), expect_user_id.get())
}
