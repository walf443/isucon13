use crate::repos::theme_repository::ThemeRepositoryInfra;
use isupipe_core::db::{DBPool, HaveDBPool};
use isupipe_core::repos::theme_repository::HaveThemeRepository;
use isupipe_core::services::theme_service::ThemeServiceImpl;
use std::sync::Arc;

pub struct ThemeServiceInfra {
    db_pool: Arc<DBPool>,
    theme_repo: ThemeRepositoryInfra,
}

impl ThemeServiceInfra {
    pub fn new(db_pool: Arc<DBPool>) -> Self {
        Self {
            db_pool,
            theme_repo: ThemeRepositoryInfra {},
        }
    }
}

impl HaveDBPool for ThemeServiceInfra {
    fn get_db_pool(&self) -> &DBPool {
        &self.db_pool
    }
}

impl HaveThemeRepository for ThemeServiceInfra {
    type Repo = ThemeRepositoryInfra;

    fn theme_repo(&self) -> &Self::Repo {
        &self.theme_repo
    }
}

impl ThemeServiceImpl for ThemeServiceInfra {}
