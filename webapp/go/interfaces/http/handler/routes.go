package handler

import (
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

// Usecases は handler が使う usecase をまとめたもの。
type Usecases struct {
	Tag                   usecase.TagUsecase
	Theme                 usecase.ThemeUsecase
	User                  usecase.UserUsecase
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
	user := newUserHandler(u.User)
	icon := newIconHandler(u.Icon, fallbackImagePath)
	livestream := newLivestreamHandler(u.Livestream, u.LivestreamReservation)
	reaction := newReactionHandler(u.Reaction)
	livecomment := newLivecommentHandler(u.Livecomment, u.LivecommentReport)
	ngWord := newNGWordHandler(u.NGWord)
	viewer := newLivestreamViewerHandler(u.LivestreamViewer)
	statistics := newStatisticsHandler(u.Statistics)
	payment := newPaymentHandler(u.Payment)
	initialize := newInitializeHandler(u.Initialize)

	// 初期化
	e.POST("/api/initialize", initialize.Initialize)

	// top
	e.GET("/api/tag", tag.GetTags)
	e.GET("/api/user/:username/theme", theme.GetStreamerTheme)

	// livestream
	// reserve livestream
	e.POST("/api/livestream/reservation", livestream.ReserveLivestream)
	// list livestream
	e.GET("/api/livestream/search", livestream.SearchLivestreams)
	e.GET("/api/livestream", livestream.GetMyLivestreams)
	e.GET("/api/user/:username/livestream", livestream.GetUserLivestreams)
	// get livestream
	e.GET("/api/livestream/:livestream_id", livestream.GetLivestream)
	// get polling livecomment timeline
	e.GET("/api/livestream/:livestream_id/livecomment", livecomment.GetLivecomments)
	// ライブコメント投稿
	e.POST("/api/livestream/:livestream_id/livecomment", livecomment.PostLivecomment)
	e.POST("/api/livestream/:livestream_id/reaction", reaction.PostReaction)
	e.GET("/api/livestream/:livestream_id/reaction", reaction.GetReactions)

	// (配信者向け)ライブコメントの報告一覧取得API
	e.GET("/api/livestream/:livestream_id/report", livecomment.GetLivecommentReports)
	e.GET("/api/livestream/:livestream_id/ngwords", ngWord.GetNGWords)
	// ライブコメント報告
	e.POST("/api/livestream/:livestream_id/livecomment/:livecomment_id/report", livecomment.PostLivecommentReport)
	// 配信者によるモデレーション (NGワード登録)
	e.POST("/api/livestream/:livestream_id/moderate", ngWord.Moderate)

	// livestream_viewersにINSERTするため必要
	// ユーザ視聴開始 (viewer)
	e.POST("/api/livestream/:livestream_id/enter", viewer.EnterLivestream)
	// ユーザ視聴終了 (viewer)
	e.DELETE("/api/livestream/:livestream_id/exit", viewer.ExitLivestream)

	// user
	e.POST("/api/register", user.Register)
	e.POST("/api/login", user.Login)
	e.GET("/api/user/me", user.GetMe)
	// フロントエンドで、配信予約のコラボレーターを指定する際に必要
	e.GET("/api/user/:username", user.GetUser)
	e.GET("/api/user/:username/statistics", statistics.GetUserStatistics)
	e.GET("/api/user/:username/icon", icon.GetIcon)
	e.POST("/api/icon", icon.PostIcon)

	// stats
	// ライブ配信統計情報
	e.GET("/api/livestream/:livestream_id/statistics", statistics.GetLivestreamStatistics)

	// 課金情報
	e.GET("/api/payment", payment.GetPaymentResult)
}
