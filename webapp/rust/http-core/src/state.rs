use sqlx::MySqlPool;
use std::sync::Arc;

#[derive(Clone)]
pub struct AppState {
    pub pool: MySqlPool,
    pub key: axum_extra::extract::cookie::Key,
    pub powerdns_subdomain_address: Arc<String>,
}

impl axum::extract::FromRef<AppState> for axum_extra::extract::cookie::Key {
    fn from_ref(state: &AppState) -> Self {
        state.key.clone()
    }
}
