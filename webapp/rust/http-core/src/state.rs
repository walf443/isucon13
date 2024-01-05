use isupipe_core::services::manager::ServiceManager;
use std::sync::Arc;

#[derive(Clone)]
pub struct AppState<S: ServiceManager> {
    pub service: S,
    pub key: axum_extra::extract::cookie::Key,
    pub powerdns_subdomain_address: Arc<String>,
}
impl<S: ServiceManager> axum::extract::FromRef<AppState<S>> for axum_extra::extract::cookie::Key {
    fn from_ref(state: &AppState<S>) -> Self {
        state.key.clone()
    }
}
