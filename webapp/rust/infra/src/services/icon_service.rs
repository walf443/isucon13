use crate::repos::icon_repository::IconRepositoryInfra;
use crate::repos::user_repository::UserRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::icon_repository::HaveIconRepository;
use isupipe_core::repos::user_repository::HaveUserRepository;
use isupipe_core::services::icon_service::IconServiceImpl;

#[derive(Clone)]
pub struct IconServiceInfra {
    db_pool: DBPool,
    icon_repo: IconRepositoryInfra,
    user_repo: UserRepositoryInfra,
}

impl IconServiceInfra {
    pub fn new(db_pool: DBPool) -> Self {
        Self {
            db_pool,
            icon_repo: IconRepositoryInfra {},
            user_repo: UserRepositoryInfra {},
        }
    }
}

impl HaveDBPool for IconServiceInfra {
    fn get_db_pool(&self) -> &DBPool {
        &self.db_pool
    }
}

impl HaveIconRepository for IconServiceInfra {
    type Repo = IconRepositoryInfra;

    fn icon_repo(&self) -> &Self::Repo {
        &self.icon_repo
    }
}

impl HaveUserRepository for IconServiceInfra {
    type Repo = UserRepositoryInfra;

    fn user_repo(&self) -> &Self::Repo {
        &self.user_repo
    }
}

impl IconServiceImpl for IconServiceInfra {}
