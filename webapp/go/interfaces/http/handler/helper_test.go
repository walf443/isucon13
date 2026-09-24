package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/sessions"
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
