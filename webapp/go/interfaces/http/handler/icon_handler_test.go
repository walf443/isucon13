package handler

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase"
)

type fakeIconUsecase struct {
	image []byte
	err   error

	iconID    domain.IconID
	updateErr error

	gotUsername string
	gotUserID   domain.UserID
	gotImage    []byte
}

func (u *fakeIconUsecase) FindImageByUsername(ctx context.Context, username string) ([]byte, error) {
	u.gotUsername = username
	return u.image, u.err
}

func (u *fakeIconUsecase) Update(ctx context.Context, userID domain.UserID, image []byte) (domain.IconID, error) {
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
			wantBody: errorBody(http.StatusNotFound, "not found user that has the given username"),
		},
		{
			name:     "returns 500 on unexpected error",
			usecase:  &fakeIconUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
			wantBody: errorBody(http.StatusInternalServerError, "boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// セッション不要のエンドポイントなので Cookie は付けない
			rec := serve(t, newIconHandler(tt.usecase, fallbackPath).GetIcon, testRequest{method: http.MethodGet, route: "/api/user/:username/icon", path: "/api/user/alice/icon"})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
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
			cookie: sessionAs(42),
			// "new icon" を base64 エンコードしたもの
			body:     `{"image":"bmV3IGljb24="}`,
			usecase:  &fakeIconUsecase{iconID: 100},
			wantCode: http.StatusCreated,
			wantBody: `{"id":100}` + "\n",
		},
		{
			name:     "returns 400 on invalid json",
			cookie:   sessionAs(42),
			body:     `{`,
			usecase:  &fakeIconUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: errorBody(http.StatusBadRequest, "failed to decode the request body as json"),
		},
		{
			name:     "returns 500 on unexpected error",
			cookie:   sessionAs(42),
			body:     `{"image":"bmV3IGljb24="}`,
			usecase:  &fakeIconUsecase{updateErr: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
			wantBody: errorBody(http.StatusInternalServerError, "boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newIconHandler(tt.usecase, "").PostIcon, testRequest{method: http.MethodPost, route: "/api/icon", path: "/api/icon", body: tt.body, cookie: tt.cookie})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
			if tt.wantCode == http.StatusCreated {
				if tt.usecase.gotUserID != 42 || string(tt.usecase.gotImage) != "new icon" {
					t.Errorf("userID = %d, image = %q", tt.usecase.gotUserID, tt.usecase.gotImage)
				}
			}
		})
	}
}
