package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
)

type fakeThemeUsecase struct {
	theme       *model.ThemeModel
	err         error
	gotUsername string
}

func (u *fakeThemeUsecase) FindByUsername(ctx context.Context, username string) (*model.ThemeModel, error) {
	u.gotUsername = username
	return u.theme, u.err
}

func TestThemeHandler_GetStreamerTheme(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 1, time.Now().Add(time.Hour))
	}

	tests := []struct {
		name     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeThemeUsecase
		wantCode int
		wantBody string
	}{
		{
			name:     "returns theme",
			cookie:   validCookie,
			usecase:  &fakeThemeUsecase{theme: &model.ThemeModel{ID: 10, UserID: 1, DarkMode: true}},
			wantCode: http.StatusOK,
			wantBody: `{"id":10,"dark_mode":true}` + "\n",
		},
		{
			// Cookie が無くても空のセッションが返るので、EXPIRES 不在として 403 になる (移行前と同じ挙動)
			name:     "returns 403 without session",
			cookie:   nil,
			usecase:  &fakeThemeUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name: "returns 401 when session has expired",
			cookie: func(t *testing.T) *http.Cookie {
				return newSessionCookie(t, 1, time.Now().Add(-time.Hour))
			},
			usecase:  &fakeThemeUsecase{},
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "returns 404 when user is not found",
			cookie:   validCookie,
			usecase:  &fakeThemeUsecase{err: usecase.ErrUserNotFound},
			wantCode: http.StatusNotFound,
		},
		{
			name:     "returns 500 on unexpected error",
			cookie:   validCookie,
			usecase:  &fakeThemeUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/user/:username/theme", newThemeHandler(tt.usecase).GetStreamerTheme)

			req := httptest.NewRequest(http.MethodGet, "/api/user/alice/theme", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie(t))
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("code = %d, want %d (body: %s)", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
			if tt.wantCode == http.StatusOK && tt.usecase.gotUsername != "alice" {
				t.Errorf("username = %q, want %q", tt.usecase.gotUsername, "alice")
			}
		})
	}
}
