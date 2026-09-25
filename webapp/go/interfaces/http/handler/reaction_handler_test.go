package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type fakeReactionUsecase struct {
	reactions []*domain.Reaction
	reaction  *domain.Reaction
	err       error

	gotLivestreamID domain.LivestreamID
	gotLimit        *domain.Limit
	gotUserID       domain.UserID
	gotEmojiName    string
}

func (u *fakeReactionUsecase) FindAllByLivestreamID(ctx context.Context, livestreamID domain.LivestreamID, limit *domain.Limit) ([]*domain.Reaction, error) {
	u.gotLivestreamID = livestreamID
	u.gotLimit = limit
	return u.reactions, u.err
}

func (u *fakeReactionUsecase) Create(ctx context.Context, userID domain.UserID, livestreamID domain.LivestreamID, emojiName string) (*domain.Reaction, error) {
	u.gotUserID = userID
	u.gotLivestreamID = livestreamID
	u.gotEmojiName = emojiName
	return u.reaction, u.err
}

var (
	testReaction = &domain.Reaction{
		ID:        100,
		EmojiName: "tada",
		User:      domain.User{ID: 1, Name: "bob", Theme: domain.ThemeModel{ID: 11, UserID: 1}, IconHash: "bbb"},
		Livestream: domain.Livestream{
			ID:    10,
			Owner: domain.User{ID: 2, Name: "alice", Theme: domain.ThemeModel{ID: 12, UserID: 2, DarkMode: true}, IconHash: "aaa"},
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
	tests := []struct {
		name      string
		path      string
		cookie    func(t *testing.T) *http.Cookie
		usecase   *fakeReactionUsecase
		wantCode  int
		wantBody  string
		wantLimit *domain.Limit
	}{
		{
			name:     "returns reactions",
			path:     "/api/livestream/10/reaction",
			cookie:   sessionAs(1),
			usecase:  &fakeReactionUsecase{reactions: []*domain.Reaction{testReaction}},
			wantCode: http.StatusOK,
			wantBody: "[" + testReactionJSON + "]\n",
		},
		{
			name:      "passes limit",
			path:      "/api/livestream/10/reaction?limit=5",
			cookie:    sessionAs(1),
			usecase:   &fakeReactionUsecase{},
			wantCode:  http.StatusOK,
			wantBody:  "[]\n",
			wantLimit: func() *domain.Limit { l := domain.Limit(5); return &l }(),
		},
		{
			name:     "returns 400 when livestream_id is not integer",
			path:     "/api/livestream/abc/reaction",
			cookie:   sessionAs(1),
			usecase:  &fakeReactionUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: errorBody(http.StatusBadRequest, "livestream_id in path must be integer"),
		},
		{
			name:     "returns 500 on unexpected error",
			path:     "/api/livestream/10/reaction",
			cookie:   sessionAs(1),
			usecase:  &fakeReactionUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
			wantBody: errorBody(http.StatusInternalServerError, "boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newReactionHandler(tt.usecase).GetReactions, testRequest{method: http.MethodGet, route: "/api/livestream/:livestream_id/reaction", path: tt.path, cookie: tt.cookie})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
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
			cookie:   sessionAs(1),
			body:     `{"emoji_name":"tada"}`,
			usecase:  &fakeReactionUsecase{reaction: testReaction},
			wantCode: http.StatusCreated,
			wantBody: testReactionJSON + "\n",
		},
		{
			name:     "returns 400 when livestream_id is not integer",
			path:     "/api/livestream/abc/reaction",
			cookie:   sessionAs(1),
			body:     `{"emoji_name":"tada"}`,
			usecase:  &fakeReactionUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: errorBody(http.StatusBadRequest, "livestream_id in path must be integer"),
		},
		{
			name:     "returns 400 on invalid json",
			path:     "/api/livestream/10/reaction",
			cookie:   sessionAs(1),
			body:     `{`,
			usecase:  &fakeReactionUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: errorBody(http.StatusBadRequest, "failed to decode the request body as json"),
		},
		{
			name:     "returns 500 on unexpected error",
			path:     "/api/livestream/10/reaction",
			cookie:   sessionAs(1),
			body:     `{"emoji_name":"tada"}`,
			usecase:  &fakeReactionUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
			wantBody: errorBody(http.StatusInternalServerError, "boom"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newReactionHandler(tt.usecase).PostReaction, testRequest{method: http.MethodPost, route: "/api/livestream/:livestream_id/reaction", path: tt.path, body: tt.body, cookie: tt.cookie})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
			if tt.wantCode == http.StatusCreated {
				if tt.usecase.gotUserID != 1 || tt.usecase.gotLivestreamID != 10 || tt.usecase.gotEmojiName != "tada" {
					t.Errorf("userID = %d, livestreamID = %d, emoji = %q", tt.usecase.gotUserID, tt.usecase.gotLivestreamID, tt.usecase.gotEmojiName)
				}
			}
		})
	}
}

func TestReactionHandler_GetReactions_Limit(t *testing.T) {
	testLimitQueryParam(t, maxReactionsLimit, func(t *testing.T, limit string) (*httptest.ResponseRecorder, *domain.Limit) {
		u := &fakeReactionUsecase{}
		rec := serve(t, newReactionHandler(u).GetReactions, testRequest{method: http.MethodGet, route: "/api/livestream/:livestream_id/reaction", path: "/api/livestream/10/reaction?limit=" + limit, cookie: sessionAs(1)})
		return rec, u.gotLimit
	})
}
