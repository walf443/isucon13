package handler

import (
	"time"

	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

// Usecases は handler が使う usecase をまとめたもの。
type Usecases struct {
	Tag                   usecase.TagUsecase
	Theme                 usecase.ThemeUsecase
	User                  usecase.UserUsecase
	UserRegistration      usecase.UserRegistrationUsecase
	Icon                  usecase.IconUsecase
	Livestream            usecase.LivestreamUsecase
	LivestreamReservation usecase.LivestreamReservationUsecase
	Reaction              usecase.ReactionUsecase
	Livecomment           usecase.LivecommentUsecase
	LivecommentReport     usecase.LivecommentReportUsecase
	NGWord                usecase.NGWordUsecase
	LivestreamViewer      usecase.LivestreamViewerUsecase
	Statistics            usecase.StatisticsUsecase
	Payment               usecase.PaymentUsecase
	Initialize            usecase.InitializeUsecase
}

// RegisterRoutes は handler を組み立てて e にルーティングを登録する。
// fallbackImagePath はアイコン未登録のユーザに返す画像ファイルのパス。
func RegisterRoutes(e *echo.Echo, u Usecases, fallbackImagePath string) {
	tag := newTagHandler(u.Tag)
	theme := newThemeHandler(u.Theme)
	user := newUserHandler(u.User, u.UserRegistration)
	icon := newIconHandler(u.Icon, fallbackImagePath)
	livestream := newLivestreamHandler(u.Livestream, u.LivestreamReservation)
	reaction := newReactionHandler(u.Reaction)
	livecomment := newLivecommentHandler(u.Livecomment, u.LivecommentReport)
	ngWord := newNGWordHandler(u.NGWord)
	viewer := newLivestreamViewerHandler(u.LivestreamViewer)
	statistics := newStatisticsHandler(u.Statistics)
	payment := newPaymentHandler(u.Payment)
	initialize := newInitializeHandler(u.Initialize)

	// ログイン済みのセッションが必要な API に付ける (付けない API はセッション無しで使える)
	auth := requireSession(time.Now)
	authWithLog := requireSessionWithLog(time.Now)

	// 初期化
	e.POST("/api/initialize", initialize.Initialize)

	// top
	e.GET("/api/tag", tag.GetTags)
	e.GET("/api/user/:username/theme", theme.GetStreamerTheme, authWithLog)

	// livestream
	// reserve livestream
	e.POST("/api/livestream/reservation", livestream.ReserveLivestream, auth)
	// list livestream
	e.GET("/api/livestream/search", livestream.SearchLivestreams)
	e.GET("/api/livestream", livestream.GetMyLivestreams, auth)
	e.GET("/api/user/:username/livestream", livestream.GetUserLivestreams, auth)
	// get livestream
	e.GET("/api/livestream/:livestream_id", livestream.GetLivestream, auth)
	// get polling livecomment timeline
	e.GET("/api/livestream/:livestream_id/livecomment", livecomment.GetLivecomments, auth)
	// ライブコメント投稿
	e.POST("/api/livestream/:livestream_id/livecomment", livecomment.PostLivecomment, auth)
	e.POST("/api/livestream/:livestream_id/reaction", reaction.PostReaction, auth)
	e.GET("/api/livestream/:livestream_id/reaction", reaction.GetReactions, auth)

	// (配信者向け)ライブコメントの報告一覧取得API
	e.GET("/api/livestream/:livestream_id/report", livecomment.GetLivecommentReports, auth)
	e.GET("/api/livestream/:livestream_id/ngwords", ngWord.GetNGWords, auth)
	// ライブコメント報告
	e.POST("/api/livestream/:livestream_id/livecomment/:livecomment_id/report", livecomment.PostLivecommentReport, auth)
	// 配信者によるモデレーション (NGワード登録)
	e.POST("/api/livestream/:livestream_id/moderate", ngWord.Moderate, auth)

	// livestream_viewersにINSERTするため必要
	// ユーザ視聴開始 (viewer)
	e.POST("/api/livestream/:livestream_id/enter", viewer.EnterLivestream, auth)
	// ユーザ視聴終了 (viewer)
	e.DELETE("/api/livestream/:livestream_id/exit", viewer.ExitLivestream, auth)

	// user
	e.POST("/api/register", user.Register)
	e.POST("/api/login", user.Login)
	e.GET("/api/user/me", user.GetMe, auth)
	// フロントエンドで、配信予約のコラボレーターを指定する際に必要
	e.GET("/api/user/:username", user.GetUser, auth)
	e.GET("/api/user/:username/statistics", statistics.GetUserStatistics, auth)
	e.GET("/api/user/:username/icon", icon.GetIcon)
	e.POST("/api/icon", icon.PostIcon, auth)

	// stats
	// ライブ配信統計情報
	e.GET("/api/livestream/:livestream_id/statistics", statistics.GetLivestreamStatistics, auth)

	// 課金情報
	e.GET("/api/payment", payment.GetPaymentResult)
}
