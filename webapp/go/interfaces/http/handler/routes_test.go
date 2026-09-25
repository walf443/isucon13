package handler

import (
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

// セッションが無い場合に、公開 API 以外は 403 を返し、公開 API は 403 を返さないことを確認する。
func TestRegisterRoutes_RequireSession(t *testing.T) {
	// セッション無しで使える API (移行前と同じ)
	public := map[string]bool{
		"POST /api/initialize":         true,
		"GET /api/tag":                 true,
		"GET /api/livestream/search":   true,
		"POST /api/register":           true,
		"POST /api/login":              true,
		"GET /api/user/:username/icon": true,
		"GET /api/payment":             true,
	}
	// パスパラメータは妥当な値にする (不正な値だと先に 400 を返す API がある)
	pathParams := strings.NewReplacer(":livestream_id", "1", ":livecomment_id", "1", ":username", "alice")

	e := newTestEcho()
	// usecase は nil なので、セッションの検証を通り抜けると panic する。テストを止めずに 500 として扱う
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{DisablePrintStack: true}))
	RegisterRoutes(e, Usecases{}, "")

	seen := map[string]bool{}
	for _, r := range e.Routes() {
		key := r.Method + " " + r.Path
		seen[key] = true
		t.Run(key, func(t *testing.T) {
			rec := send(t, e, testRequest{method: r.Method, path: pathParams.Replace(r.Path)})
			if public[key] {
				if rec.Code == http.StatusForbidden {
					t.Errorf("public route returned 403 (body: %s)", rec.Body.String())
				}
				return
			}
			assertResponse(t, rec, http.StatusForbidden, `{"message":"failed to get EXPIRES value from session"}`+"\n")
		})
	}
	for key := range public {
		if !seen[key] {
			t.Errorf("public route %s is not registered", key)
		}
	}
}
