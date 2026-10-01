package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase"
)

type fakeThemeUsecase struct {
	theme       *domain.Theme
	err         error
	gotUsername domain.Username
}

func (u *fakeThemeUsecase) FindByUsername(ctx context.Context, username domain.Username) (*domain.Theme, error) {
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
			usecase:  &fakeThemeUsecase{theme: &domain.Theme{ID: 10, UserID: 1, DarkMode: true}},
			wantCode: http.StatusOK,
			wantBody: `{"id":10,"dark_mode":true}` + "\n",
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

// 不正な形のユーザ名のユーザは存在しないので、ユーザが見つからない場合と同じ応答になり、usecase は呼ばれない。
func TestThemeHandler_GetStreamerTheme_InvalidUsername(t *testing.T) {
	u := &fakeThemeUsecase{}
	rec := serve(t, newThemeHandler(u).GetStreamerTheme, testRequest{method: http.MethodGet, route: "/api/user/:username/theme", path: "/api/user/a.b/theme", cookie: sessionAs(1)})
	assertResponse(t, rec, http.StatusNotFound, errorBody(http.StatusNotFound, "not found user that has the given username"))
	if u.gotUsername != "" {
		t.Errorf("usecase was called with %q, want it not to be called", u.gotUsername)
	}
}
