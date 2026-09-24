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

type fakeLivestreamUsecase struct {
	livestream  *model.Livestream
	livestreams []*model.Livestream
	err         error

	gotID     int64
	gotUserID int64
}

func (u *fakeLivestreamUsecase) FindAllByUserID(ctx context.Context, userID int64) ([]*model.Livestream, error) {
	u.gotUserID = userID
	return u.livestreams, u.err
}

func (u *fakeLivestreamUsecase) FindByID(ctx context.Context, id int64) (*model.Livestream, error) {
	u.gotID = id
	return u.livestream, u.err
}

func TestLivestreamHandler_GetLivestream(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 1, time.Now().Add(time.Hour))
	}
	owner := model.User{
		ID:          2,
		Name:        "alice",
		DisplayName: "Alice",
		Description: "hello",
		Theme:       model.ThemeModel{ID: 20, UserID: 2, DarkMode: true},
		IconHash:    "abc",
	}

	tests := []struct {
		name     string
		path     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeLivestreamUsecase
		wantCode int
		wantBody string
	}{
		{
			name:   "returns livestream",
			path:   "/api/livestream/10",
			cookie: validCookie,
			usecase: &fakeLivestreamUsecase{livestream: &model.Livestream{
				ID:           10,
				Owner:        owner,
				Title:        "stream",
				Description:  "desc",
				PlaylistUrl:  "https://example.com/p.m3u8",
				ThumbnailUrl: "https://example.com/t.jpg",
				Tags:         []model.TagModel{{ID: 1, Name: "ゲーム実況"}},
				StartAt:      1700000000,
				EndAt:        1700003600,
			}},
			wantCode: http.StatusOK,
			wantBody: `{"id":10,"owner":{"id":2,"name":"alice","display_name":"Alice","description":"hello","theme":{"id":20,"dark_mode":true},"icon_hash":"abc"},"title":"stream","description":"desc","playlist_url":"https://example.com/p.m3u8","thumbnail_url":"https://example.com/t.jpg","tags":[{"id":1,"name":"ゲーム実況"}],"start_at":1700000000,"end_at":1700003600}` + "\n",
		},
		{
			// タグが無い場合も null ではなく [] を返す (移行前と同じ)
			name:     "returns empty tags as array",
			path:     "/api/livestream/10",
			cookie:   validCookie,
			usecase:  &fakeLivestreamUsecase{livestream: &model.Livestream{ID: 10, Owner: owner}},
			wantCode: http.StatusOK,
			wantBody: `{"id":10,"owner":{"id":2,"name":"alice","display_name":"Alice","description":"hello","theme":{"id":20,"dark_mode":true},"icon_hash":"abc"},"title":"","description":"","playlist_url":"","thumbnail_url":"","tags":[],"start_at":0,"end_at":0}` + "\n",
		},
		{
			name:     "returns 403 without session",
			path:     "/api/livestream/10",
			cookie:   nil,
			usecase:  &fakeLivestreamUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 400 when livestream_id is not integer",
			path:     "/api/livestream/abc",
			cookie:   validCookie,
			usecase:  &fakeLivestreamUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when livestream is not found",
			path:     "/api/livestream/10",
			cookie:   validCookie,
			usecase:  &fakeLivestreamUsecase{err: usecase.ErrLivestreamNotFound},
			wantCode: http.StatusNotFound,
		},
		{
			name:     "returns 500 on unexpected error",
			path:     "/api/livestream/10",
			cookie:   validCookie,
			usecase:  &fakeLivestreamUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/livestream/:livestream_id", NewLivestreamHandler(tt.usecase).GetLivestream)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie(t))
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("code = %d, want %d (body: %s)", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %s\nwant   %s", rec.Body.String(), tt.wantBody)
			}
			if tt.wantCode == http.StatusOK && tt.usecase.gotID != 10 {
				t.Errorf("id = %d, want 10", tt.usecase.gotID)
			}
		})
	}
}

func TestLivestreamHandler_GetMyLivestreams(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 42, time.Now().Add(time.Hour))
	}
	owner := model.User{ID: 42, Name: "alice", Theme: model.ThemeModel{ID: 20, UserID: 42}, IconHash: "abc"}
	ownerJSON := `{"id":42,"name":"alice","theme":{"id":20,"dark_mode":false},"icon_hash":"abc"}`

	tests := []struct {
		name     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeLivestreamUsecase
		wantCode int
		wantBody string
	}{
		{
			name:   "returns livestreams of logged-in user",
			cookie: validCookie,
			usecase: &fakeLivestreamUsecase{livestreams: []*model.Livestream{
				{ID: 1, Owner: owner, Title: "s1", Tags: []model.TagModel{{ID: 1, Name: "t1"}}},
				{ID: 2, Owner: owner, Title: "s2"},
			}},
			wantCode: http.StatusOK,
			wantBody: `[` +
				`{"id":1,"owner":` + ownerJSON + `,"title":"s1","description":"","playlist_url":"","thumbnail_url":"","tags":[{"id":1,"name":"t1"}],"start_at":0,"end_at":0},` +
				`{"id":2,"owner":` + ownerJSON + `,"title":"s2","description":"","playlist_url":"","thumbnail_url":"","tags":[],"start_at":0,"end_at":0}` +
				`]` + "\n",
		},
		{
			// 配信が無い場合も null ではなく [] を返す (移行前と同じ)
			name:     "returns empty array when no livestreams",
			cookie:   validCookie,
			usecase:  &fakeLivestreamUsecase{livestreams: nil},
			wantCode: http.StatusOK,
			wantBody: "[]\n",
		},
		{
			name:     "returns 403 without session",
			cookie:   nil,
			usecase:  &fakeLivestreamUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 500 on unexpected error",
			cookie:   validCookie,
			usecase:  &fakeLivestreamUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/livestream", NewLivestreamHandler(tt.usecase).GetMyLivestreams)

			req := httptest.NewRequest(http.MethodGet, "/api/livestream", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie(t))
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("code = %d, want %d (body: %s)", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %s\nwant   %s", rec.Body.String(), tt.wantBody)
			}
			if tt.wantCode == http.StatusOK && tt.usecase.gotUserID != 42 {
				t.Errorf("userID = %d, want 42", tt.usecase.gotUserID)
			}
		})
	}
}
