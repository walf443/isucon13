package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
)

type fakeIconUsecase struct {
	image []byte
	err   error

	iconID    model.IconID
	updateErr error

	gotUsername string
	gotUserID   model.UserID
	gotImage    []byte
}

func (u *fakeIconUsecase) FindImageByUsername(ctx context.Context, username string) ([]byte, error) {
	u.gotUsername = username
	return u.image, u.err
}

func (u *fakeIconUsecase) Update(ctx context.Context, userID model.UserID, image []byte) (model.IconID, error) {
	u.gotUserID = userID
	u.gotImage = image
	return u.iconID, u.updateErr
}

func TestIconHandler_GetIcon(t *testing.T) {
	fallbackPath := filepath.Join(t.TempDir(), "NoImage.jpg")
	if err := os.WriteFile(fallbackPath, []byte("fallback image"), 0o644); err != nil {
		t.Fatalf("failed to write fallback image: %v", err)
	}

	tests := []struct {
		name            string
		usecase         *fakeIconUsecase
		wantCode        int
		wantBody        string
		wantContentType string
	}{
		{
			name:            "returns registered icon",
			usecase:         &fakeIconUsecase{image: []byte("icon")},
			wantCode:        http.StatusOK,
			wantBody:        "icon",
			wantContentType: "image/jpeg",
		},
		{
			name:            "returns fallback image when icon is not registered",
			usecase:         &fakeIconUsecase{err: usecase.ErrIconNotFound},
			wantCode:        http.StatusOK,
			wantBody:        "fallback image",
			wantContentType: "image/jpeg",
		},
		{
			name:     "returns 404 when user is not found",
			usecase:  &fakeIconUsecase{err: usecase.ErrUserNotFound},
			wantCode: http.StatusNotFound,
		},
		{
			name:     "returns 500 on unexpected error",
			usecase:  &fakeIconUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/user/:username/icon", newIconHandler(tt.usecase, fallbackPath).GetIcon)

			// セッション不要のエンドポイントなので Cookie は付けない
			req := httptest.NewRequest(http.MethodGet, "/api/user/alice/icon", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("code = %d, want %d (body: %s)", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
			if tt.wantContentType != "" && rec.Header().Get("Content-Type") != tt.wantContentType {
				t.Errorf("Content-Type = %q, want %q", rec.Header().Get("Content-Type"), tt.wantContentType)
			}
			if tt.usecase.gotUsername != "alice" {
				t.Errorf("username = %q, want %q", tt.usecase.gotUsername, "alice")
			}
		})
	}
}

func TestIconHandler_PostIcon(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 42, time.Now().Add(time.Hour))
	}

	tests := []struct {
		name     string
		cookie   func(t *testing.T) *http.Cookie
		body     string
		usecase  *fakeIconUsecase
		wantCode int
		wantBody string
	}{
		{
			name:   "updates icon",
			cookie: validCookie,
			// "new icon" を base64 エンコードしたもの
			body:     `{"image":"bmV3IGljb24="}`,
			usecase:  &fakeIconUsecase{iconID: 100},
			wantCode: http.StatusCreated,
			wantBody: `{"id":100}` + "\n",
		},
		{
			name:     "returns 403 without session",
			cookie:   nil,
			body:     `{"image":"bmV3IGljb24="}`,
			usecase:  &fakeIconUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 400 on invalid json",
			cookie:   validCookie,
			body:     `{`,
			usecase:  &fakeIconUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on unexpected error",
			cookie:   validCookie,
			body:     `{"image":"bmV3IGljb24="}`,
			usecase:  &fakeIconUsecase{updateErr: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.POST("/api/icon", newIconHandler(tt.usecase, "").PostIcon)

			req := httptest.NewRequest(http.MethodPost, "/api/icon", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
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
			if tt.wantCode == http.StatusCreated {
				if tt.usecase.gotUserID != 42 || string(tt.usecase.gotImage) != "new icon" {
					t.Errorf("userID = %d, image = %q", tt.usecase.gotUserID, tt.usecase.gotImage)
				}
			}
		})
	}
}
