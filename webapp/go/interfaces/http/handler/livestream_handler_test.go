package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
)

type fakeLivestreamUsecase struct {
	livestream  *model.Livestream
	livestreams []*model.Livestream
	err         error

	gotID       model.LivestreamID
	gotUserID   model.UserID
	gotUsername string
	gotTagName  string
	gotLimit    *model.Limit
	gotInput    usecase.ReserveLivestreamInput
	// calls は呼ばれたメソッド名を順に記録する
	calls []string
}

func (u *fakeLivestreamUsecase) Reserve(ctx context.Context, userID model.UserID, input usecase.ReserveLivestreamInput) (*model.Livestream, error) {
	u.calls = append(u.calls, "Reserve")
	u.gotUserID = userID
	u.gotInput = input
	return u.livestream, u.err
}

func (u *fakeLivestreamUsecase) FindAllByTagName(ctx context.Context, tagName string) ([]*model.Livestream, error) {
	u.calls = append(u.calls, "FindAllByTagName")
	u.gotTagName = tagName
	return u.livestreams, u.err
}

func (u *fakeLivestreamUsecase) FindAll(ctx context.Context, limit *model.Limit) ([]*model.Livestream, error) {
	u.calls = append(u.calls, "FindAll")
	u.gotLimit = limit
	return u.livestreams, u.err
}

func (u *fakeLivestreamUsecase) FindAllByUsername(ctx context.Context, username string) ([]*model.Livestream, error) {
	u.gotUsername = username
	return u.livestreams, u.err
}

func (u *fakeLivestreamUsecase) FindAllByUserID(ctx context.Context, userID model.UserID) ([]*model.Livestream, error) {
	u.gotUserID = userID
	return u.livestreams, u.err
}

func (u *fakeLivestreamUsecase) FindByID(ctx context.Context, id model.LivestreamID) (*model.Livestream, error) {
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
			e.GET("/api/livestream/:livestream_id", newLivestreamHandler(tt.usecase).GetLivestream)

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
			e.GET("/api/livestream", newLivestreamHandler(tt.usecase).GetMyLivestreams)

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

func TestLivestreamHandler_GetUserLivestreams(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 1, time.Now().Add(time.Hour))
	}
	owner := model.User{ID: 42, Name: "alice", Theme: model.ThemeModel{ID: 20, UserID: 42}, IconHash: "abc"}

	tests := []struct {
		name     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeLivestreamUsecase
		wantCode int
		wantBody string
	}{
		{
			name:     "returns livestreams of the user",
			cookie:   validCookie,
			usecase:  &fakeLivestreamUsecase{livestreams: []*model.Livestream{{ID: 1, Owner: owner, Title: "s1"}}},
			wantCode: http.StatusOK,
			wantBody: `[{"id":1,"owner":{"id":42,"name":"alice","theme":{"id":20,"dark_mode":false},"icon_hash":"abc"},"title":"s1","description":"","playlist_url":"","thumbnail_url":"","tags":[],"start_at":0,"end_at":0}]` + "\n",
		},
		{
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
			name:     "returns 404 when user is not found",
			cookie:   validCookie,
			usecase:  &fakeLivestreamUsecase{err: usecase.ErrUserNotFound},
			wantCode: http.StatusNotFound,
			// このエンドポイントだけ 404 のメッセージが他と異なるので確認しておく (echo のデフォルトのエラーハンドラの形式)
			wantBody: `{"message":"user not found"}` + "\n",
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
			e.GET("/api/user/:username/livestream", newLivestreamHandler(tt.usecase).GetUserLivestreams)

			req := httptest.NewRequest(http.MethodGet, "/api/user/alice/livestream", nil)
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
			if tt.wantCode == http.StatusOK && tt.usecase.gotUsername != "alice" {
				t.Errorf("username = %q, want %q", tt.usecase.gotUsername, "alice")
			}
		})
	}
}

func TestLivestreamHandler_SearchLivestreams(t *testing.T) {
	owner := model.User{ID: 42, Name: "alice", Theme: model.ThemeModel{ID: 20, UserID: 42}, IconHash: "abc"}
	found := []*model.Livestream{{ID: 1, Owner: owner, Title: "s1"}}
	foundJSON := `[{"id":1,"owner":{"id":42,"name":"alice","theme":{"id":20,"dark_mode":false},"icon_hash":"abc"},"title":"s1","description":"","playlist_url":"","thumbnail_url":"","tags":[],"start_at":0,"end_at":0}]` + "\n"
	five := model.Limit(5)

	tests := []struct {
		name        string
		query       string
		usecase     *fakeLivestreamUsecase
		wantCode    int
		wantBody    string
		wantCalls   []string
		wantTagName string
		wantLimit   *model.Limit
	}{
		{
			name:        "searches by tag",
			query:       "?tag=%E3%82%B2%E3%83%BC%E3%83%A0",
			usecase:     &fakeLivestreamUsecase{livestreams: found},
			wantCode:    http.StatusOK,
			wantBody:    foundJSON,
			wantCalls:   []string{"FindAllByTagName"},
			wantTagName: "ゲーム",
		},
		{
			// tag がある場合 limit は見ない (不正な値でも 400 にならない)
			name:        "ignores limit when tag is given",
			query:       "?tag=foo&limit=abc",
			usecase:     &fakeLivestreamUsecase{livestreams: found},
			wantCode:    http.StatusOK,
			wantCalls:   []string{"FindAllByTagName"},
			wantTagName: "foo",
		},
		{
			name:      "lists all without condition",
			query:     "",
			usecase:   &fakeLivestreamUsecase{livestreams: found},
			wantCode:  http.StatusOK,
			wantBody:  foundJSON,
			wantCalls: []string{"FindAll"},
		},
		{
			name:      "lists with limit",
			query:     "?limit=5",
			usecase:   &fakeLivestreamUsecase{livestreams: found},
			wantCode:  http.StatusOK,
			wantCalls: []string{"FindAll"},
			wantLimit: &five,
		},
		{
			name:        "returns empty array when nothing found",
			query:       "?tag=nothing",
			usecase:     &fakeLivestreamUsecase{livestreams: []*model.Livestream{}},
			wantCode:    http.StatusOK,
			wantBody:    "[]\n",
			wantCalls:   []string{"FindAllByTagName"},
			wantTagName: "nothing",
		},
		{
			name:     "returns 500 on unexpected error",
			query:    "",
			usecase:  &fakeLivestreamUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/livestream/search", newLivestreamHandler(tt.usecase).SearchLivestreams)

			// セッション不要のエンドポイントなので Cookie は付けない
			req := httptest.NewRequest(http.MethodGet, "/api/livestream/search"+tt.query, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("code = %d, want %d (body: %s)", rec.Code, tt.wantCode, rec.Body.String())
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %s\nwant   %s", rec.Body.String(), tt.wantBody)
			}
			if tt.wantCalls != nil && !slices.Equal(tt.usecase.calls, tt.wantCalls) {
				t.Errorf("calls = %v, want %v", tt.usecase.calls, tt.wantCalls)
			}
			if tt.usecase.gotTagName != tt.wantTagName {
				t.Errorf("tag name = %q, want %q", tt.usecase.gotTagName, tt.wantTagName)
			}
			if (tt.usecase.gotLimit == nil) != (tt.wantLimit == nil) || (tt.wantLimit != nil && *tt.usecase.gotLimit != *tt.wantLimit) {
				t.Errorf("limit = %v, want %v", tt.usecase.gotLimit, tt.wantLimit)
			}
		})
	}
}

func TestLivestreamHandler_SearchLivestreams_Limit(t *testing.T) {
	testLimitQueryParam(t, maxLivestreamsLimit, func(t *testing.T, limit string) (*httptest.ResponseRecorder, *model.Limit) {
		u := &fakeLivestreamUsecase{}
		e := newTestEcho()
		e.GET("/api/livestream/search", newLivestreamHandler(u).SearchLivestreams)

		// セッションは不要
		req := httptest.NewRequest(http.MethodGet, "/api/livestream/search?limit="+limit, nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec, u.gotLimit
	})
}

func TestLivestreamHandler_ReserveLivestream(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 2, time.Now().Add(time.Hour))
	}
	owner := model.User{ID: 2, Name: "alice", Theme: model.ThemeModel{ID: 20, UserID: 2, DarkMode: true}, IconHash: "abc"}
	reqBody := `{"tags":[1,3],"title":"stream","description":"desc","playlist_url":"https://example.com/p.m3u8","thumbnail_url":"https://example.com/t.jpg","start_at":1700874000,"end_at":1700877600}`

	tests := []struct {
		name     string
		cookie   func(t *testing.T) *http.Cookie
		body     string
		usecase  *fakeLivestreamUsecase
		wantCode int
		wantBody string
	}{
		{
			name:   "reserves livestream",
			cookie: validCookie,
			body:   reqBody,
			usecase: &fakeLivestreamUsecase{livestream: &model.Livestream{
				ID:           10,
				Owner:        owner,
				Title:        "stream",
				Description:  "desc",
				PlaylistUrl:  "https://example.com/p.m3u8",
				ThumbnailUrl: "https://example.com/t.jpg",
				Tags:         []model.TagModel{{ID: 1, Name: "ゲーム実況"}, {ID: 3, Name: "雑談"}},
				StartAt:      1700874000,
				EndAt:        1700877600,
			}},
			wantCode: http.StatusCreated,
			wantBody: `{"id":10,"owner":{"id":2,"name":"alice","theme":{"id":20,"dark_mode":true},"icon_hash":"abc"},"title":"stream","description":"desc","playlist_url":"https://example.com/p.m3u8","thumbnail_url":"https://example.com/t.jpg","tags":[{"id":1,"name":"ゲーム実況"},{"id":3,"name":"雑談"}],"start_at":1700874000,"end_at":1700877600}` + "\n",
		},
		{
			name:     "returns 403 without session",
			cookie:   nil,
			body:     reqBody,
			usecase:  &fakeLivestreamUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 400 on invalid json",
			cookie:   validCookie,
			body:     `{`,
			usecase:  &fakeLivestreamUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"failed to decode the request body as json"}` + "\n",
		},
		{
			name:     "returns 400 on bad time range",
			cookie:   validCookie,
			body:     reqBody,
			usecase:  &fakeLivestreamUsecase{err: usecase.ErrBadReservationTimeRange},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"bad reservation time range"}` + "\n",
		},
		{
			// メッセージは usecase のエラーのものをそのまま返す
			name:     "returns 400 when slot is unavailable",
			cookie:   validCookie,
			body:     reqBody,
			usecase:  &fakeLivestreamUsecase{err: &usecase.ReservationSlotUnavailableError{StartAt: 1700874000, EndAt: 1700877600}},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"予約期間 1700874000 ~ 1732496400に対して、予約区間 1700874000 ~ 1700877600が予約できません"}` + "\n",
		},
		{
			name:     "returns 500 on unexpected error",
			cookie:   validCookie,
			body:     reqBody,
			usecase:  &fakeLivestreamUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"message":"boom"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.POST("/api/livestream/reservation", newLivestreamHandler(tt.usecase).ReserveLivestream)

			req := httptest.NewRequest(http.MethodPost, "/api/livestream/reservation", strings.NewReader(tt.body))
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
				t.Errorf("body = %s\nwant   %s", rec.Body.String(), tt.wantBody)
			}
			if tt.wantCode == http.StatusCreated {
				wantInput := usecase.ReserveLivestreamInput{
					TagIDs:       []model.TagID{1, 3},
					Title:        "stream",
					Description:  "desc",
					PlaylistUrl:  "https://example.com/p.m3u8",
					ThumbnailUrl: "https://example.com/t.jpg",
					StartAt:      1700874000,
					EndAt:        1700877600,
				}
				if tt.usecase.gotUserID != 2 || !reflect.DeepEqual(tt.usecase.gotInput, wantInput) {
					t.Errorf("userID = %d, input = %+v", tt.usecase.gotUserID, tt.usecase.gotInput)
				}
			}
		})
	}
}
