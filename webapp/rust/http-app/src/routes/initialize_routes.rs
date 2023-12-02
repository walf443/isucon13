use isupipe_http_core::error::Error;

#[derive(Debug, serde::Serialize)]
pub struct InitializeResponse {
    language: &'static str,
}

pub async fn initialize_handler() -> Result<axum::Json<InitializeResponse>, Error> {
    let output = tokio::process::Command::new("../sql/init.sh")
        .output()
        .await?;
    if !output.status.success() {
        return Err(Error::InternalServerError(format!(
            "init.sh failed with stdout={} stderr={}",
            String::from_utf8_lossy(&output.stdout),
            String::from_utf8_lossy(&output.stderr),
        )));
    }

    Ok(axum::Json(InitializeResponse { language: "rust" }))
}
