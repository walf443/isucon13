use crate::routes::initialize_routes::initialize_handler;
use crate::routes::livestream_routes::{get_my_livestreams_handler, livestreams_routes};
use crate::routes::login_routes::login_handler;
use crate::routes::payment_routes::get_payment_result;
use crate::routes::register_routes::register_handler;
use crate::routes::tag_routes::get_tag_handler;
use crate::routes::user_icon_routes::post_icon_handler;
use crate::routes::user_routes::user_routes;
use crate::state::AppState;
use axum::Router;
use isupipe_core::services::manager::ServiceManager;

pub mod initialize_routes;
pub mod livestream_comment_report_routes;
pub mod livestream_comment_routes;
pub mod livestream_reaction_routes;
pub mod livestream_routes;
pub mod login_routes;
pub mod payment_routes;
pub mod register_routes;
pub mod tag_routes;
pub mod user_icon_routes;
pub mod user_routes;

pub fn routes<S: ServiceManager + 'static>() -> Router<AppState<S>> {
    axum::Router::new()
        // 初期化
        .route("/api/initialize", axum::routing::post(initialize_handler))
        // top
        .route("/api/tag", axum::routing::get(get_tag_handler))
        .route("/api/register", axum::routing::post(register_handler))
        .route("/api/login", axum::routing::post(login_handler))
        .route("/api/icon", axum::routing::post(post_icon_handler))
        // 課金情報
        .route("/api/payment", axum::routing::get(get_payment_result))
        .nest("/api/user/", user_routes())
        .route(
            "/api/livestream",
            axum::routing::get(get_my_livestreams_handler),
        )
        .nest("/api/livestream/", livestreams_routes())
}
