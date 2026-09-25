package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/sessions"
	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

var testSessionStore = sessions.NewCookieStore([]byte("test-secret"))

// newTestEcho は本番と同じセッションミドルウェアを持つ echo を返す。
func newTestEcho() *echo.Echo {
	e := echo.New()
	e.Use(session.Middleware(testSessionStore))
	return e
}

// newSessionCookie は指定したユーザでログイン済みのセッション Cookie を返す。
func newSessionCookie(t *testing.T, userID int64, expires time.Time) *http.Cookie {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	sess, err := testSessionStore.Get(req, DefaultSessionIDKey)
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}
	sess.Values[DefaultSessionIDKey] = "test-session-id"
	sess.Values[DefaultUserIDKey] = userID
	sess.Values[DefaultUsernameKey] = "test-user"
	sess.Values[DefaultSessionExpiresKey] = expires.Unix()
	if err := sess.Save(req, rec); err != nil {
		t.Fatalf("failed to save session: %v", err)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("len(cookies) = %d, want 1", len(cookies))
	}
	return cookies[0]
}

// sessionAs は userID のユーザでログイン済み (有効期限は 1 時間後) の Cookie を作る関数を返す。
func sessionAs(userID int64) func(t *testing.T) *http.Cookie {
	return func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, userID, time.Now().Add(time.Hour))
	}
}

// testRequest は handler のテストで送るリクエスト。
type testRequest struct {
	method string
	// route は handler を登録するパス (例: "/api/livestream/:livestream_id")。
	route string
	// path は実際にリクエストするパス (クエリを含む)。
	path string
	// body は JSON のリクエスト本文。空ならボディを付けない。
	body string
	// cookie が nil ならセッション無しでリクエストする。
	cookie func(t *testing.T) *http.Cookie
}

// serve は h を route に登録した echo に req を送り、レスポンスを返す。
func serve(t *testing.T, h echo.HandlerFunc, req testRequest) *httptest.ResponseRecorder {
	t.Helper()
	e := newTestEcho()
	e.Add(req.method, req.route, h)
	return send(t, e, req)
}

// send は e に req を送り、レスポンスを返す。req.route は使わない。
func send(t *testing.T, e *echo.Echo, req testRequest) *httptest.ResponseRecorder {
	t.Helper()
	var httpReq *http.Request
	if req.body != "" {
		httpReq = httptest.NewRequest(req.method, req.path, strings.NewReader(req.body))
		httpReq.Header.Set("Content-Type", "application/json")
	} else {
		httpReq = httptest.NewRequest(req.method, req.path, nil)
	}
	if req.cookie != nil {
		httpReq.AddCookie(req.cookie(t))
	}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httpReq)
	return rec
}

// assertResponse はステータスコードと、wantBody が空でなければ本文を確認する。
// ステータスコードが違う場合はテストを止める。
func assertResponse(t *testing.T, rec *httptest.ResponseRecorder, wantCode int, wantBody string) {
	t.Helper()
	if rec.Code != wantCode {
		t.Fatalf("code = %d, want %d (body: %s)", rec.Code, wantCode, rec.Body.String())
	}
	if wantBody != "" && rec.Body.String() != wantBody {
		t.Errorf("body = %s\nwant   %s", rec.Body.String(), wantBody)
	}
}

// testLimitQueryParam は limit クエリパラメータの検証を、境界値を含めて確認する。
// max はその API の limit の上限。do は limit クエリに値を付けてリクエストし、レスポンスと usecase に渡った limit を返す。
func testLimitQueryParam(t *testing.T, max model.Limit, do func(t *testing.T, limit string) (*httptest.ResponseRecorder, *model.Limit)) {
	t.Helper()
	outOfRange := fmt.Sprintf(`{"message":"limit query parameter must be between 1 and %d"}`, max) + "\n"
	notInteger := `{"message":"limit query parameter must be integer"}` + "\n"
	maxStr := strconv.FormatInt(int64(max), 10)
	overMaxStr := strconv.FormatInt(int64(max)+1, 10)

	tests := []struct {
		limit     string
		wantCode  int
		wantBody  string
		wantLimit model.Limit
	}{
		{limit: "1", wantCode: http.StatusOK, wantLimit: 1},
		{limit: maxStr, wantCode: http.StatusOK, wantLimit: max},
		// 0 件の取得や負の数、上限を超える値は 400 (移行前は 0 は空配列、負の数は 500、上限なし)
		{limit: "0", wantCode: http.StatusBadRequest, wantBody: outOfRange},
		{limit: "-1", wantCode: http.StatusBadRequest, wantBody: outOfRange},
		{limit: overMaxStr, wantCode: http.StatusBadRequest, wantBody: outOfRange},
		{limit: "abc", wantCode: http.StatusBadRequest, wantBody: notInteger},
	}
	for _, tt := range tests {
		t.Run("limit="+tt.limit, func(t *testing.T) {
			rec, gotLimit := do(t, tt.limit)
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
			if tt.wantCode == http.StatusOK && (gotLimit == nil || *gotLimit != tt.wantLimit) {
				t.Errorf("limit = %v, want %d", gotLimit, tt.wantLimit)
			}
			if tt.wantCode != http.StatusOK && gotLimit != nil {
				t.Errorf("usecase was called with limit %d, want not called", *gotLimit)
			}
		})
	}
}
