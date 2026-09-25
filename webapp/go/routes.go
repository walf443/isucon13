package main

import "github.com/labstack/echo/v4"

func registerRoutes(e *echo.Echo, h *handlers) {
	// 初期化
	e.POST("/api/initialize", initializeHandler)

	// top
	e.GET("/api/tag", h.tag.GetTags)
	e.GET("/api/user/:username/theme", h.theme.GetStreamerTheme)

	// livestream
	// reserve livestream
	e.POST("/api/livestream/reservation", reserveLivestreamHandler)
	// list livestream
	e.GET("/api/livestream/search", h.livestream.SearchLivestreams)
	e.GET("/api/livestream", h.livestream.GetMyLivestreams)
	e.GET("/api/user/:username/livestream", h.livestream.GetUserLivestreams)
	// get livestream
	e.GET("/api/livestream/:livestream_id", h.livestream.GetLivestream)
	// get polling livecomment timeline
	e.GET("/api/livestream/:livestream_id/livecomment", h.livecomment.GetLivecomments)
	// ライブコメント投稿
	e.POST("/api/livestream/:livestream_id/livecomment", h.livecomment.PostLivecomment)
	e.POST("/api/livestream/:livestream_id/reaction", h.reaction.PostReaction)
	e.GET("/api/livestream/:livestream_id/reaction", h.reaction.GetReactions)

	// (配信者向け)ライブコメントの報告一覧取得API
	e.GET("/api/livestream/:livestream_id/report", h.livecomment.GetLivecommentReports)
	e.GET("/api/livestream/:livestream_id/ngwords", h.ngWord.GetNGWords)
	// ライブコメント報告
	e.POST("/api/livestream/:livestream_id/livecomment/:livecomment_id/report", reportLivecommentHandler)
	// 配信者によるモデレーション (NGワード登録)
	e.POST("/api/livestream/:livestream_id/moderate", h.ngWord.Moderate)

	// livestream_viewersにINSERTするため必要
	// ユーザ視聴開始 (viewer)
	e.POST("/api/livestream/:livestream_id/enter", enterLivestreamHandler)
	// ユーザ視聴終了 (viewer)
	e.DELETE("/api/livestream/:livestream_id/exit", exitLivestreamHandler)

	// user
	e.POST("/api/register", registerHandler)
	e.POST("/api/login", loginHandler)
	e.GET("/api/user/me", h.user.GetMe)
	// フロントエンドで、配信予約のコラボレーターを指定する際に必要
	e.GET("/api/user/:username", h.user.GetUser)
	e.GET("/api/user/:username/statistics", getUserStatisticsHandler)
	e.GET("/api/user/:username/icon", h.icon.GetIcon)
	e.POST("/api/icon", h.icon.PostIcon)

	// stats
	// ライブ配信統計情報
	e.GET("/api/livestream/:livestream_id/statistics", getLivestreamStatisticsHandler)

	// 課金情報
	e.GET("/api/payment", GetPaymentResult)
}
