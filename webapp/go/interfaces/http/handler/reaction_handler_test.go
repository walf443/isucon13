package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type fakeReactionUsecase struct {
	reactions []*model.Reaction
	reaction  *model.Reaction
	err       error

	gotLivestreamID model.LivestreamID
	gotLimit        *model.Limit
	gotUserID       model.UserID
	gotEmojiName    string
}

func (u *fakeReactionUsecase) FindAllByLivestreamID(ctx context.Context, livestreamID model.LivestreamID, limit *model.Limit) ([]*model.Reaction, error) {
	u.gotLivestreamID = livestreamID
	u.gotLimit = limit
	return u.reactions, u.err
}

func (u *fakeReactionUsecase) Create(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID, emojiName string) (*model.Reaction, error) {
	u.gotUserID = userID
	u.gotLivestreamID = livestreamID
	u.gotEmojiName = emojiName
	return u.reaction, u.err
}

var (
	testReaction = &model.Reaction{
		ID:        100,
		EmojiName: "tada",
		User:      model.User{ID: 1, Name: "bob", Theme: model.ThemeModel{ID: 11, UserID: 1}, IconHash: "bbb"},
		Livestream: model.Livestream{
			ID:    10,
			Owner: model.User{ID: 2, Name: "alice", Theme: model.ThemeModel{ID: 12, UserID: 2, DarkMode: true}, IconHash: "aaa"},
			Title: "stream",
		},
		CreatedAt: 1700000000,
	}
	testReactionJSON = `{"id":100,"emoji_name":"tada",` +
		`"user":{"id":1,"name":"bob","theme":{"id":11,"dark_mode":false},"icon_hash":"bbb"},` +
		`"livestream":{"id":10,"owner":{"id":2,"name":"alice","theme":{"id":12,"dark_mode":true},"icon_hash":"aaa"},"title":"stream","description":"","playlist_url":"","thumbnail_url":"","tags":[],"start_at":0,"end_at":0},` +
		`"created_at":1700000000}`
)

func TestReactionHandler_GetReactions(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 1, time.Now().Add(time.Hour))
	}

	tests := []struct {
		name      string
		path      string
		cookie    func(t *testing.T) *http.Cookie
		usecase   *fakeReactionUsecase
		wantCode  int
		wantBody  string
		wantLimit *model.Limit
	}{
		{
			name:     "returns reactions",
			path:     "/api/livestream/10/reaction",
			cookie:   validCookie,
			usecase:  &fakeReactionUsecase{reactions: []*model.Reaction{testReaction}},
			wantCode: http.StatusOK,
			wantBody: "[" + testReactionJSON + "]\n",
		},
		{
			name:      "passes limit",
			path:      "/api/livestream/10/reaction?limit=5",
			cookie:    validCookie,
			usecase:   &fakeReactionUsecase{},
			wantCode:  http.StatusOK,
			wantBody:  "[]\n",
			wantLimit: func() *model.Limit { l := model.Limit(5); return &l }(),
		},
		{
			name:     "returns 403 without session",
			path:     "/api/livestream/10/reaction",
			cookie:   nil,
			usecase:  &fakeReactionUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 400 when livestream_id is not integer",
			path:     "/api/livestream/abc/reaction",
			cookie:   validCookie,
			usecase:  &fakeReactionUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "returns 400 when limit is not integer",
			path:     "/api/livestream/10/reaction?limit=abc",
			cookie:   validCookie,
			usecase:  &fakeReactionUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"limit query parameter must be integer"}` + "\n",
		},
		{
			// 0 件の取得や負の数、上限を超える値は 400 (移行前は 0 は空配列、負の数は 500、上限なし)
			name:     "returns 400 when limit is zero",
			path:     "/api/livestream/10/reaction?limit=0",
			cookie:   validCookie,
			usecase:  &fakeReactionUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"limit query parameter must be between 1 and 100"}` + "\n",
		},
		{
			name:     "returns 400 when limit is negative",
			path:     "/api/livestream/10/reaction?limit=-1",
			cookie:   validCookie,
			usecase:  &fakeReactionUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"limit query parameter must be between 1 and 100"}` + "\n",
		},
		{
			name:     "returns 400 when limit is over the max",
			path:     "/api/livestream/10/reaction?limit=101",
			cookie:   validCookie,
			usecase:  &fakeReactionUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"limit query parameter must be between 1 and 100"}` + "\n",
		},
		{
			name:     "returns 500 on unexpected error",
			path:     "/api/livestream/10/reaction",
			cookie:   validCookie,
			usecase:  &fakeReactionUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/livestream/:livestream_id/reaction", NewReactionHandler(tt.usecase).GetReactions)

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
			if tt.wantCode == http.StatusOK {
				if tt.usecase.gotLivestreamID != 10 {
					t.Errorf("livestreamID = %d, want 10", tt.usecase.gotLivestreamID)
				}
				if (tt.usecase.gotLimit == nil) != (tt.wantLimit == nil) || (tt.wantLimit != nil && *tt.usecase.gotLimit != *tt.wantLimit) {
					t.Errorf("limit = %v, want %v", tt.usecase.gotLimit, tt.wantLimit)
				}
			}
		})
	}
}

func TestReactionHandler_PostReaction(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 1, time.Now().Add(time.Hour))
	}

	tests := []struct {
		name     string
		path     string
		cookie   func(t *testing.T) *http.Cookie
		body     string
		usecase  *fakeReactionUsecase
		wantCode int
		wantBody string
	}{
		{
			name:     "creates reaction",
			path:     "/api/livestream/10/reaction",
			cookie:   validCookie,
			body:     `{"emoji_name":"tada"}`,
			usecase:  &fakeReactionUsecase{reaction: testReaction},
			wantCode: http.StatusCreated,
			wantBody: testReactionJSON + "\n",
		},
		{
			// livestream_id のチェックはセッション検証より先 (移行前と同じ)
			name:     "returns 400 for invalid livestream_id even without session",
			path:     "/api/livestream/abc/reaction",
			cookie:   nil,
			body:     `{"emoji_name":"tada"}`,
			usecase:  &fakeReactionUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "returns 403 without session",
			path:     "/api/livestream/10/reaction",
			cookie:   nil,
			body:     `{"emoji_name":"tada"}`,
			usecase:  &fakeReactionUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 400 on invalid json",
			path:     "/api/livestream/10/reaction",
			cookie:   validCookie,
			body:     `{`,
			usecase:  &fakeReactionUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on unexpected error",
			path:     "/api/livestream/10/reaction",
			cookie:   validCookie,
			body:     `{"emoji_name":"tada"}`,
			usecase:  &fakeReactionUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.POST("/api/livestream/:livestream_id/reaction", NewReactionHandler(tt.usecase).PostReaction)

			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
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
				if tt.usecase.gotUserID != 1 || tt.usecase.gotLivestreamID != 10 || tt.usecase.gotEmojiName != "tada" {
					t.Errorf("userID = %d, livestreamID = %d, emoji = %q", tt.usecase.gotUserID, tt.usecase.gotLivestreamID, tt.usecase.gotEmojiName)
				}
			}
		})
	}
}
