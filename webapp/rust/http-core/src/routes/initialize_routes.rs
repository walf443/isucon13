use crate::error::Error;
use crate::state::AppState;
use axum::extract::State;
use isupipe_core::services::initialize_service::InitializeService;
use isupipe_core::services::manager::ServiceManager;

#[derive(Debug, serde::Serialize)]
pub struct InitializeResponse {
    language: &'static str,
}

pub async fn initialize_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
) -> Result<axum::Json<InitializeResponse>, Error> {
    let output = service.initialize_service().execute_command().await?;

    if !output.success {
        return Err(Error::InternalServerError(format!(
            "init.sh failed with stdout={} stderr={}",
            String::from_utf8_lossy(&output.stdout),
            String::from_utf8_lossy(&output.stderr),
        )));
    }

    Ok(axum::Json(InitializeResponse { language: "rust" }))
}
