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

type fakeUserUsecase struct {
	user    *model.User
	err     error
	gotID   int64
	gotName string
}

func (u *fakeUserUsecase) FindByID(ctx context.Context, id int64) (*model.User, error) {
	u.gotID = id
	return u.user, u.err
}

func (u *fakeUserUsecase) FindByName(ctx context.Context, name string) (*model.User, error) {
	u.gotName = name
	return u.user, u.err
}

func TestUserHandler_GetUser(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 1, time.Now().Add(time.Hour))
	}

	tests := []struct {
		name     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeUserUsecase
		wantCode int
		wantBody string
	}{
		{
			name:   "returns user",
			cookie: validCookie,
			usecase: &fakeUserUsecase{user: &model.User{
				ID:          1,
				Name:        "alice",
				DisplayName: "Alice",
				Description: "hello",
				Theme:       model.ThemeModel{ID: 10, UserID: 1, DarkMode: true},
				IconHash:    "abc",
			}},
			wantCode: http.StatusOK,
			wantBody: `{"id":1,"name":"alice","display_name":"Alice","description":"hello","theme":{"id":10,"dark_mode":true},"icon_hash":"abc"}` + "\n",
		},
		{
			name:     "returns 403 without session",
			cookie:   nil,
			usecase:  &fakeUserUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 404 when user is not found",
			cookie:   validCookie,
			usecase:  &fakeUserUsecase{err: usecase.ErrUserNotFound},
			wantCode: http.StatusNotFound,
		},
		{
			name:     "returns 500 on unexpected error",
			cookie:   validCookie,
			usecase:  &fakeUserUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/user/:username", NewUserHandler(tt.usecase).GetUser)

			req := httptest.NewRequest(http.MethodGet, "/api/user/alice", nil)
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
			if tt.wantCode == http.StatusOK && tt.usecase.gotName != "alice" {
				t.Errorf("name = %q, want %q", tt.usecase.gotName, "alice")
			}
		})
	}
}

func TestUserHandler_GetMe(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 42, time.Now().Add(time.Hour))
	}

	tests := []struct {
		name     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeUserUsecase
		wantCode int
		wantBody string
	}{
		{
			name:   "returns logged-in user",
			cookie: validCookie,
			usecase: &fakeUserUsecase{user: &model.User{
				ID:          42,
				Name:        "alice",
				DisplayName: "Alice",
				Description: "hello",
				Theme:       model.ThemeModel{ID: 10, UserID: 42, DarkMode: false},
				IconHash:    "abc",
			}},
			wantCode: http.StatusOK,
			wantBody: `{"id":42,"name":"alice","display_name":"Alice","description":"hello","theme":{"id":10,"dark_mode":false},"icon_hash":"abc"}` + "\n",
		},
		{
			name:     "returns 403 without session",
			cookie:   nil,
			usecase:  &fakeUserUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 404 when session user is not found",
			cookie:   validCookie,
			usecase:  &fakeUserUsecase{err: usecase.ErrUserNotFound},
			wantCode: http.StatusNotFound,
		},
		{
			name:     "returns 500 on unexpected error",
			cookie:   validCookie,
			usecase:  &fakeUserUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/user/me", NewUserHandler(tt.usecase).GetMe)

			req := httptest.NewRequest(http.MethodGet, "/api/user/me", nil)
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
			if tt.wantCode == http.StatusOK && tt.usecase.gotID != 42 {
				t.Errorf("id = %d, want 42", tt.usecase.gotID)
			}
		})
	}
}
