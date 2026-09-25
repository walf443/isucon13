package handler

import (
	"context"
	"errors"
	"net/http"
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
	tests := []struct {
		name     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeThemeUsecase
		wantCode int
		wantBody string
	}{
		{
			name:     "returns theme",
			cookie:   sessionAs(1),
			usecase:  &fakeThemeUsecase{theme: &model.ThemeModel{ID: 10, UserID: 1, DarkMode: true}},
			wantCode: http.StatusOK,
			wantBody: `{"id":10,"dark_mode":true}` + "\n",
		},
		{
			name: "returns 401 when session has expired",
			cookie: func(t *testing.T) *http.Cookie {
				return newSessionCookie(t, 1, time.Now().Add(-time.Hour))
			},
			usecase:  &fakeThemeUsecase{},
			wantCode: http.StatusUnauthorized,
			wantBody: errorBody(http.StatusUnauthorized, "session has expired"),
		},
		{
			name:     "returns 404 when user is not found",
			cookie:   sessionAs(1),
			usecase:  &fakeThemeUsecase{err: usecase.ErrUserNotFound},
			wantCode: http.StatusNotFound,
			wantBody: errorBody(http.StatusNotFound, "not found user that has the given username"),
		},
		{
			name:     "returns 500 on unexpected error",
			cookie:   sessionAs(1),
			usecase:  &fakeThemeUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
			wantBody: errorBody(http.StatusInternalServerError, "boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newThemeHandler(tt.usecase).GetStreamerTheme, testRequest{method: http.MethodGet, route: "/api/user/:username/theme", path: "/api/user/alice/theme", cookie: tt.cookie})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
			if tt.wantCode == http.StatusOK && tt.usecase.gotUsername != "alice" {
				t.Errorf("username = %q, want %q", tt.usecase.gotUsername, "alice")
			}
		})
	}
}
