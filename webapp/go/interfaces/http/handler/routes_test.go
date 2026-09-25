package handler

import (
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

// 移行前 (package main の routes.go) と同じルーティングになっていることを確認する。
func TestRegisterRoutes(t *testing.T) {
	want := []struct {
		method  string
		path    string
		handler string
	}{
		{http.MethodPost, "/api/initialize", "(*initializeHandler).Initialize"},
		{http.MethodGet, "/api/tag", "(*tagHandler).GetTags"},
		{http.MethodGet, "/api/user/:username/theme", "(*themeHandler).GetStreamerTheme"},
		{http.MethodPost, "/api/livestream/reservation", "(*livestreamHandler).ReserveLivestream"},
		{http.MethodGet, "/api/livestream/search", "(*livestreamHandler).SearchLivestreams"},
		{http.MethodGet, "/api/livestream", "(*livestreamHandler).GetMyLivestreams"},
		{http.MethodGet, "/api/user/:username/livestream", "(*livestreamHandler).GetUserLivestreams"},
		{http.MethodGet, "/api/livestream/:livestream_id", "(*livestreamHandler).GetLivestream"},
		{http.MethodGet, "/api/livestream/:livestream_id/livecomment", "(*livecommentHandler).GetLivecomments"},
		{http.MethodPost, "/api/livestream/:livestream_id/livecomment", "(*livecommentHandler).PostLivecomment"},
		{http.MethodPost, "/api/livestream/:livestream_id/reaction", "(*reactionHandler).PostReaction"},
		{http.MethodGet, "/api/livestream/:livestream_id/reaction", "(*reactionHandler).GetReactions"},
		{http.MethodGet, "/api/livestream/:livestream_id/report", "(*livecommentHandler).GetLivecommentReports"},
		{http.MethodGet, "/api/livestream/:livestream_id/ngwords", "(*ngWordHandler).GetNGWords"},
		{http.MethodPost, "/api/livestream/:livestream_id/livecomment/:livecomment_id/report", "(*livecommentHandler).PostLivecommentReport"},
		{http.MethodPost, "/api/livestream/:livestream_id/moderate", "(*ngWordHandler).Moderate"},
		{http.MethodPost, "/api/livestream/:livestream_id/enter", "(*livestreamViewerHandler).EnterLivestream"},
		{http.MethodDelete, "/api/livestream/:livestream_id/exit", "(*livestreamViewerHandler).ExitLivestream"},
		{http.MethodPost, "/api/register", "(*userHandler).Register"},
		{http.MethodPost, "/api/login", "(*userHandler).Login"},
		{http.MethodGet, "/api/user/me", "(*userHandler).GetMe"},
		{http.MethodGet, "/api/user/:username", "(*userHandler).GetUser"},
		{http.MethodGet, "/api/user/:username/statistics", "(*statisticsHandler).GetUserStatistics"},
		{http.MethodGet, "/api/user/:username/icon", "(*iconHandler).GetIcon"},
		{http.MethodPost, "/api/icon", "(*iconHandler).PostIcon"},
		{http.MethodGet, "/api/livestream/:livestream_id/statistics", "(*statisticsHandler).GetLivestreamStatistics"},
		{http.MethodGet, "/api/payment", "(*paymentHandler).GetPaymentResult"},
	}

	e := echo.New()
	RegisterRoutes(e, Usecases{}, "")

	got := map[string]string{}
	for _, r := range e.Routes() {
		got[r.Method+" "+r.Path] = r.Name
	}
	if len(got) != len(want) {
		keys := make([]string, 0, len(got))
		for k := range got {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		t.Errorf("len(routes) = %d, want %d: %v", len(got), len(want), keys)
	}
	for _, w := range want {
		name, ok := got[w.method+" "+w.path]
		if !ok {
			t.Errorf("%s %s is not registered", w.method, w.path)
			continue
		}
		// Name は "<パッケージ>.(*xxxHandler).Method-fm" の形式
		if !strings.HasSuffix(name, "."+w.handler+"-fm") {
			t.Errorf("%s %s is handled by %s, want %s", w.method, w.path, name, w.handler)
		}
	}
}
