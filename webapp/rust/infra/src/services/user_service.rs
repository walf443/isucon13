use crate::repos::user_repository::UserRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::user_repository::HaveUserRepository;
use isupipe_core::services::user_service::UserServiceImpl;

#[derive(Clone)]
pub struct UserServiceInfra {
    db_pool: DBPool,
    user_repo: UserRepositoryInfra,
}

impl UserServiceInfra {
    pub fn new(db_pool: DBPool) -> Self {
        Self {
            db_pool,
            user_repo: UserRepositoryInfra {},
        }
    }
}

impl HaveDBPool for UserServiceInfra {
    fn get_db_pool(&self) -> &DBPool {
        &self.db_pool
    }
}

impl HaveUserRepository for UserServiceInfra {
    type Repo = UserRepositoryInfra;

    fn user_repo(&self) -> &Self::Repo {
        &self.user_repo
    }
}

impl UserServiceImpl for UserServiceInfra {}
