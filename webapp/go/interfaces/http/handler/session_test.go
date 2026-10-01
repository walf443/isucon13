package handler

import (
	"bytes"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestRequireSession(t *testing.T) {
	now := time.Unix(1700000000, 0)

	tests := []struct {
		name     string
		cookie   func(t *testing.T) *http.Cookie
		wantCode int
		wantBody string
	}{
		{
			name:     "passes when session expires at now",
			cookie:   func(t *testing.T) *http.Cookie { return newSessionCookie(t, 1, now) },
			wantCode: http.StatusOK,
			wantBody: "ok",
		},
		{
			name:     "returns 401 when session expired 1 second ago",
			cookie:   func(t *testing.T) *http.Cookie { return newSessionCookie(t, 1, now.Add(-time.Second)) },
			wantCode: http.StatusUnauthorized,
			wantBody: errorBody(http.StatusUnauthorized, "session has expired"),
		},
		{
			name:     "returns 403 without session",
			cookie:   nil,
			wantCode: http.StatusForbidden,
			wantBody: errorBody(http.StatusForbidden, "failed to get EXPIRES value from session"),
		},
		{
			name: "returns 401 when session has no user id",
			cookie: func(t *testing.T) *http.Cookie {
				return newSessionCookieWithValues(t, map[any]any{defaultSessionExpiresKey: now.Add(time.Hour).Unix()})
			},
			wantCode: http.StatusUnauthorized,
			wantBody: errorBody(http.StatusUnauthorized, "failed to get USERID value from session"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			h := func(c echo.Context) error {
				called = true
				return c.String(http.StatusOK, "ok")
			}
			e := newTestEcho()
			e.GET("/", h, requireSession(func() time.Time { return now }))

			rec := send(t, e, testRequest{method: http.MethodGet, path: "/", cookie: tt.cookie})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
			if wantCalled := tt.wantCode == http.StatusOK; called != wantCalled {
				t.Errorf("handler called = %v, want %v", called, wantCalled)
			}
		})
	}
}

// 移行前の GET /api/user/:username/theme と同じく、検証に失敗したときだけログを出すことを確認する。
func TestRequireSessionWithLog(t *testing.T) {
	now := time.Unix(1700000000, 0)

	tests := []struct {
		name    string
		cookie  func(t *testing.T) *http.Cookie
		wantLog string
	}{
		{
			name:    "logs verification error",
			cookie:  nil,
			wantLog: "verifyUserSession: code=403, message=failed to get EXPIRES value from session",
		},
		{
			name:   "does not log on success",
			cookie: func(t *testing.T) *http.Cookie { return newSessionCookie(t, 1, now) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			var logs bytes.Buffer
			e.Logger.SetOutput(&logs)
			e.GET("/", func(c echo.Context) error { return c.NoContent(http.StatusOK) }, requireSessionWithLog(func() time.Time { return now }))

			send(t, e, testRequest{method: http.MethodGet, path: "/", cookie: tt.cookie})
			// errorResponseHandler もログを出すので、このメッセージの有無だけを確認する
			gotLog := strings.Contains(logs.String(), "verifyUserSession: ")
			if tt.wantLog == "" && gotLog {
				t.Errorf("unexpected log: %s", logs.String())
			}
			if tt.wantLog != "" && !strings.Contains(logs.String(), tt.wantLog) {
				t.Errorf("log = %q, want to contain %q", logs.String(), tt.wantLog)
			}
		})
	}
}
