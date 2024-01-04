use async_session::{CookieStore, SessionStore};
use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum_extra::extract::SignedCookieJar;
use chrono::{DateTime, NaiveDate, TimeZone, Utc};
use isupipe_core::models::livestream::{CreateLivestream, Livestream, LivestreamId};
use isupipe_core::models::livestream_ranking_entry::LivestreamRankingEntry;
use isupipe_core::models::livestream_statistics::LivestreamStatistics;
use isupipe_core::models::livestream_viewers_history::CreateLivestreamViewersHistory;
use isupipe_core::models::ng_word::{CreateNgWord, NgWord};
use isupipe_core::models::tag::TagId;
use isupipe_core::models::user::UserId;
use isupipe_core::repos::livestream_comment_report_repository::LivestreamCommentReportRepository;
use isupipe_core::repos::livestream_comment_repository::LivestreamCommentRepository;
use isupipe_core::repos::livestream_repository::LivestreamRepository;
use isupipe_core::repos::livestream_tag_repository::LivestreamTagRepository;
use isupipe_core::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepository;
use isupipe_core::repos::reaction_repository::ReactionRepository;
use isupipe_core::repos::reservation_slot_repository::ReservationSlotRepository;
use isupipe_core::services::livestream_service::LivestreamService;
use isupipe_core::services::livestream_viewers_history_service::LivestreamViewersHistoryService;
use isupipe_core::services::manager::ServiceManager;
use isupipe_core::services::ng_word_service::NgWordService;
use isupipe_http_core::error::Error;
use isupipe_http_core::responses::livestream_response::LivestreamResponse;
use isupipe_http_core::state::AppState;
use isupipe_http_core::{verify_user_session, DEFAULT_SESSION_ID_KEY, DEFAULT_USER_ID_KEY};
use isupipe_infra::repos::livestream_comment_report_repository::LivestreamCommentReportRepositoryInfra;
use isupipe_infra::repos::livestream_comment_repository::LivestreamCommentRepositoryInfra;
use isupipe_infra::repos::livestream_repository::LivestreamRepositoryInfra;
use isupipe_infra::repos::livestream_tag_repository::LivestreamTagRepositoryInfra;
use isupipe_infra::repos::livestream_viewers_history_repository::LivestreamViewersHistoryRepositoryInfra;
use isupipe_infra::repos::reaction_repository::ReactionRepositoryInfra;
use isupipe_infra::repos::reservation_slot_repository::ReservationSlotRepositoryInfra;

#[derive(Debug, serde::Deserialize)]
pub struct ReserveLivestreamRequest {
    tags: Vec<i64>,
    title: String,
    description: String,
    playlist_url: String,
    thumbnail_url: String,
    start_at: i64,
    end_at: i64,
}

pub async fn reserve_livestream_handler<S: ServiceManager>(
    State(AppState { service, pool, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    axum::Json(req): axum::Json<ReserveLivestreamRequest>,
) -> Result<(StatusCode, axum::Json<LivestreamResponse>), Error> {
    verify_user_session(&jar).await?;

    if req.tags.iter().any(|&tag_id| tag_id > 103) {
        tracing::error!("unexpected tags: {:?}", req);
    }

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);

    let mut tx = pool.begin().await?;

    // 2023/11/25 10:00からの１年間の期間内であるかチェック
    let term_start_at = Utc.from_utc_datetime(
        &NaiveDate::from_ymd_opt(2023, 11, 25)
            .unwrap()
            .and_hms_opt(1, 0, 0)
            .unwrap(),
    );
    let term_end_at = Utc.from_utc_datetime(
        &NaiveDate::from_ymd_opt(2024, 11, 25)
            .unwrap()
            .and_hms_opt(1, 0, 0)
            .unwrap(),
    );
    let reserve_start_at = DateTime::from_timestamp(req.start_at, 0).unwrap();
    let reserve_end_at = DateTime::from_timestamp(req.end_at, 0).unwrap();
    if reserve_start_at >= term_end_at || reserve_end_at <= term_start_at {
        return Err(Error::BadRequest("bad reservation time range".into()));
    }

    let reservation_slot_repo = ReservationSlotRepositoryInfra {};

    // 予約枠をみて、予約が可能か調べる
    // NOTE: 並列な予約のoverbooking防止にFOR UPDATEが必要
    let slots = reservation_slot_repo
        .find_all_between_for_update(&mut tx, req.start_at, req.end_at)
        .await
        .map_err(|e| {
            tracing::warn!("予約枠一覧取得でエラー発生: {e:?}");
            e
        })?;

    for slot in slots {
        let count = reservation_slot_repo
            .find_slot_between(&mut tx, slot.start_at, slot.end_at)
            .await?;
        tracing::info!(
            "{} ~ {}予約枠の残数 = {}",
            slot.start_at,
            slot.end_at,
            slot.slot
        );
        if count < 1 {
            return Err(Error::BadRequest(
                format!(
                    "予約期間 {} ~ {}に対して、予約区間 {} ~ {}が予約できません",
                    term_start_at.timestamp(),
                    term_end_at.timestamp(),
                    req.start_at,
                    req.end_at
                )
                .into(),
            ));
        }
    }

    reservation_slot_repo
        .decrement_slot_between(&mut tx, req.start_at, req.end_at)
        .await?;

    let livestream_repo = LivestreamRepositoryInfra {};
    let livestream_id = livestream_repo
        .create(
            &mut tx,
            &CreateLivestream {
                user_id: user_id.clone(),
                title: req.title.clone(),
                description: req.description.clone(),
                playlist_url: req.playlist_url.clone(),
                thumbnail_url: req.thumbnail_url.clone(),
                start_at: req.start_at,
                end_at: req.end_at,
            },
        )
        .await?;

    let livestream_tag_repo = LivestreamTagRepositoryInfra {};
    // タグ追加
    for tag_id in req.tags {
        livestream_tag_repo
            .insert(&mut tx, &livestream_id, &TagId::new(tag_id))
            .await?;
    }

    let livestream = LivestreamResponse::build_by_service(
        &service,
        &Livestream {
            id: livestream_id.clone(),
            user_id: user_id.clone(),
            title: req.title,
            description: req.description,
            playlist_url: req.playlist_url,
            thumbnail_url: req.thumbnail_url,
            start_at: req.start_at,
            end_at: req.end_at,
        },
    )
    .await?;

    tx.commit().await?;

    Ok((StatusCode::CREATED, axum::Json(livestream)))
}

#[derive(Debug, serde::Deserialize)]
pub struct SearchLivestreamsQuery {
    #[serde(default)]
    pub tag: String,
    #[serde(default)]
    pub limit: String,
}

pub async fn search_livestreams_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    Query(SearchLivestreamsQuery {
        tag: key_tag_name,
        limit,
    }): Query<SearchLivestreamsQuery>,
) -> Result<axum::Json<Vec<LivestreamResponse>>, Error> {
    let livestream_models: Vec<Livestream> = if key_tag_name.is_empty() {
        let limit = if limit.is_empty() {
            None
        } else {
            let limit: i64 = limit
                .parse()
                .map_err(|_| Error::BadRequest("failed to parse limit".into()))?;

            Some(limit)
        };

        service
            .livestream_service()
            .find_recent_livestreams(limit)
            .await?
    } else {
        // タグによる取得
        service
            .livestream_service()
            .find_recent_by_tag_name(&key_tag_name)
            .await?
    };

    let livestreams =
        LivestreamResponse::bulk_build_by_service(&service, &livestream_models).await?;

    Ok(axum::Json(livestreams))
}
pub async fn get_my_livestreams_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
) -> Result<axum::Json<Vec<LivestreamResponse>>, Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);

    let livestream_models = service
        .livestream_service()
        .find_all_by_user_id(&user_id)
        .await?;

    let livestreams =
        LivestreamResponse::bulk_build_by_service(&service, &livestream_models).await?;

    Ok(axum::Json(livestreams))
}

pub async fn get_livestream_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<axum::Json<LivestreamResponse>, Error> {
    verify_user_session(&jar).await?;
    let livestream_id = LivestreamId::new(livestream_id);

    let livestream_model = service.livestream_service().find(&livestream_id).await?;

    if livestream_model.is_none() {
        return Err(Error::NotFound(
            "not found livestream that has the given id".into(),
        ));
    }
    let livestream_model = livestream_model.unwrap();

    let livestream = LivestreamResponse::build_by_service(&service, &livestream_model).await?;

    Ok(axum::Json(livestream))
}
pub async fn get_ngwords<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<axum::Json<Vec<NgWord>>, Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);
    let livestream_id = LivestreamId::new(livestream_id);

    let ng_words = service
        .ng_word_service()
        .find_all_by_livestream_id_and_user_id(&livestream_id, &user_id)
        .await?;

    Ok(axum::Json(ng_words))
}
#[derive(Debug, serde::Deserialize)]
pub struct ModerateRequest {
    ng_word: String,
}

#[derive(Debug, serde::Serialize)]
pub struct ModerateResponse {
    word_id: i64,
}

// NGワードを登録
pub async fn moderate_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
    axum::Json(req): axum::Json<ModerateRequest>,
) -> Result<(StatusCode, axum::Json<ModerateResponse>), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);

    let livestream_id = LivestreamId::new(livestream_id);

    // 配信者自身の配信に対するmoderateなのかを検証
    let is_exist = service
        .livestream_service()
        .exist_by_id_and_user_id(&livestream_id, &user_id)
        .await?;
    if !is_exist {
        return Err(Error::BadRequest(
            "A streamer can't moderate livestreams that other streamers own".into(),
        ));
    }

    let created_at = Utc::now().timestamp();
    let word_id = service
        .ng_word_service()
        .create(&CreateNgWord {
            user_id: user_id.clone(),
            livestream_id: livestream_id.clone(),
            word: req.ng_word,
            created_at: created_at.clone(),
        })
        .await?;

    Ok((
        StatusCode::CREATED,
        axum::Json(ModerateResponse {
            word_id: word_id.get(),
        }),
    ))
}

// viewerテーブルの廃止
pub async fn enter_livestream_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<(), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);
    let livestream_id = LivestreamId::new(livestream_id);

    let created_at = Utc::now().timestamp();

    service
        .livestream_viewers_history_service()
        .create(&CreateLivestreamViewersHistory {
            user_id: user_id.clone(),
            livestream_id: livestream_id.clone(),
            created_at: created_at.clone(),
        })
        .await?;

    Ok(())
}
pub async fn exit_livestream_handler<S: ServiceManager>(
    State(AppState { service, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<(), Error> {
    verify_user_session(&jar).await?;

    let cookie = jar.get(DEFAULT_SESSION_ID_KEY).ok_or(Error::SessionError)?;
    let sess = CookieStore::new()
        .load_session(cookie.value().to_owned())
        .await?
        .ok_or(Error::SessionError)?;
    let user_id: i64 = sess.get(DEFAULT_USER_ID_KEY).ok_or(Error::SessionError)?;
    let user_id = UserId::new(user_id);
    let livestream_id = LivestreamId::new(livestream_id);

    service
        .livestream_viewers_history_service()
        .delete_by_livestream_id_and_user_id(&livestream_id, &user_id)
        .await?;

    Ok(())
}
pub async fn get_livestream_statistics_handler<S: ServiceManager>(
    State(AppState { pool, .. }): State<AppState<S>>,
    jar: SignedCookieJar,
    Path((livestream_id,)): Path<(i64,)>,
) -> Result<axum::Json<LivestreamStatistics>, Error> {
    verify_user_session(&jar).await?;

    let livestream_id = LivestreamId::new(livestream_id);

    let mut tx = pool.begin().await?;
    let livestream_repo = LivestreamRepositoryInfra {};

    let livestream = livestream_repo
        .find(&mut tx, &livestream_id)
        .await?
        .ok_or(Error::BadRequest("".into()))?;

    let livestreams = livestream_repo.find_all(&mut tx).await?;

    // ランク算出
    let mut ranking = Vec::new();
    let reaction_repo = ReactionRepositoryInfra {};
    let comment_repo = LivestreamCommentRepositoryInfra {};
    for livestream in livestreams {
        let reactions = reaction_repo
            .count_by_livestream_id(&mut tx, &livestream.id)
            .await?;

        let total_tips = comment_repo
            .get_sum_tip_of_livestream_id(&mut tx, &livestream.id)
            .await?;

        let score = reactions + total_tips;
        ranking.push(LivestreamRankingEntry {
            livestream_id: LivestreamId::new(livestream.id.get()),
            score,
        })
    }
    ranking.sort_by(|a, b| {
        a.score
            .cmp(&b.score)
            .then_with(|| a.livestream_id.get().cmp(&b.livestream_id.get()))
    });

    let rpos = ranking
        .iter()
        .rposition(|entry| entry.livestream_id.get() == livestream_id.get())
        .unwrap();
    let rank = (ranking.len() - rpos) as i64;

    // 視聴者数算出
    let history_repo = LivestreamViewersHistoryRepositoryInfra {};
    let viewers_count = history_repo
        .count_by_livestream_id(&mut tx, &livestream_id)
        .await?;

    // 最大チップ額
    let max_tip = comment_repo
        .get_max_tip_of_livestream_id(&mut tx, &livestream.id)
        .await?;

    // リアクション数
    let reaction_repo = ReactionRepositoryInfra {};
    let total_reactions = reaction_repo
        .count_by_livestream_id(&mut tx, &livestream.id)
        .await?;

    // スパム報告数
    let report_repo = LivestreamCommentReportRepositoryInfra {};
    let total_reports = report_repo
        .count_by_livestream_id(&mut tx, &livestream.id)
        .await?;

    tx.commit().await?;

    Ok(axum::Json(LivestreamStatistics {
        rank,
        viewers_count,
        max_tip,
        total_reactions,
        total_reports,
    }))
}
