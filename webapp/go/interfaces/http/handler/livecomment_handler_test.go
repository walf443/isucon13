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

type fakeLivecommentUsecase struct {
	livecomments []*model.Livecomment
	reports      []*model.LivecommentReport
	err          error

	gotLivestreamID model.LivestreamID
	gotLimit        *int64
	gotUserID       model.UserID
}

func (u *fakeLivecommentUsecase) FindAllByLivestreamID(ctx context.Context, livestreamID model.LivestreamID, limit *int64) ([]*model.Livecomment, error) {
	u.gotLivestreamID = livestreamID
	u.gotLimit = limit
	return u.livecomments, u.err
}

func (u *fakeLivecommentUsecase) FindAllReportsByLivestreamID(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) ([]*model.LivecommentReport, error) {
	u.gotUserID = userID
	u.gotLivestreamID = livestreamID
	return u.reports, u.err
}

var (
	testLivecomment = &model.Livecomment{
		ID:   50,
		User: model.User{ID: 1, Name: "bob", Theme: model.ThemeModel{ID: 11, UserID: 1}, IconHash: "bbb"},
		Livestream: model.Livestream{
			ID:    10,
			Owner: model.User{ID: 2, Name: "alice", Theme: model.ThemeModel{ID: 12, UserID: 2, DarkMode: true}, IconHash: "aaa"},
			Title: "stream",
		},
		Comment:   "hello",
		Tip:       100,
		CreatedAt: 1700000000,
	}
	testLivecommentJSON = `{"id":50,` +
		`"user":{"id":1,"name":"bob","theme":{"id":11,"dark_mode":false},"icon_hash":"bbb"},` +
		`"livestream":{"id":10,"owner":{"id":2,"name":"alice","theme":{"id":12,"dark_mode":true},"icon_hash":"aaa"},"title":"stream","description":"","playlist_url":"","thumbnail_url":"","tags":[],"start_at":0,"end_at":0},` +
		`"comment":"hello","tip":100,"created_at":1700000000}`
)

func TestLivecommentHandler_GetLivecomments(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 1, time.Now().Add(time.Hour))
	}

	tests := []struct {
		name      string
		path      string
		cookie    func(t *testing.T) *http.Cookie
		usecase   *fakeLivecommentUsecase
		wantCode  int
		wantBody  string
		wantLimit *int64
	}{
		{
			name:     "returns livecomments",
			path:     "/api/livestream/10/livecomment",
			cookie:   validCookie,
			usecase:  &fakeLivecommentUsecase{livecomments: []*model.Livecomment{testLivecomment}},
			wantCode: http.StatusOK,
			wantBody: "[" + testLivecommentJSON + "]\n",
		},
		{
			name:      "passes limit and returns empty array",
			path:      "/api/livestream/10/livecomment?limit=5",
			cookie:    validCookie,
			usecase:   &fakeLivecommentUsecase{},
			wantCode:  http.StatusOK,
			wantBody:  "[]\n",
			wantLimit: func() *int64 { l := int64(5); return &l }(),
		},
		{
			name:     "returns 403 without session",
			path:     "/api/livestream/10/livecomment",
			cookie:   nil,
			usecase:  &fakeLivecommentUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 400 when livestream_id is not integer",
			path:     "/api/livestream/abc/livecomment",
			cookie:   validCookie,
			usecase:  &fakeLivecommentUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "returns 400 when limit is not integer",
			path:     "/api/livestream/10/livecomment?limit=abc",
			cookie:   validCookie,
			usecase:  &fakeLivecommentUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on unexpected error",
			path:     "/api/livestream/10/livecomment",
			cookie:   validCookie,
			usecase:  &fakeLivecommentUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/livestream/:livestream_id/livecomment", NewLivecommentHandler(tt.usecase).GetLivecomments)

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

func TestLivecommentHandler_GetLivecommentReports(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 2, time.Now().Add(time.Hour))
	}
	report := &model.LivecommentReport{
		ID:          7,
		Reporter:    model.User{ID: 3, Name: "carol", Theme: model.ThemeModel{ID: 13, UserID: 3}, IconHash: "ccc"},
		Livecomment: *testLivecomment,
		CreatedAt:   1700000100,
	}

	tests := []struct {
		name     string
		path     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeLivecommentUsecase
		wantCode int
		wantBody string
	}{
		{
			name:     "returns reports",
			path:     "/api/livestream/10/report",
			cookie:   validCookie,
			usecase:  &fakeLivecommentUsecase{reports: []*model.LivecommentReport{report}},
			wantCode: http.StatusOK,
			wantBody: `[{"id":7,"reporter":{"id":3,"name":"carol","theme":{"id":13,"dark_mode":false},"icon_hash":"ccc"},"livecomment":` + testLivecommentJSON + `,"created_at":1700000100}]` + "\n",
		},
		{
			name:     "returns empty array when no reports",
			path:     "/api/livestream/10/report",
			cookie:   validCookie,
			usecase:  &fakeLivecommentUsecase{},
			wantCode: http.StatusOK,
			wantBody: "[]\n",
		},
		{
			name:     "returns 403 without session",
			path:     "/api/livestream/10/report",
			cookie:   nil,
			usecase:  &fakeLivecommentUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 400 when livestream_id is not integer",
			path:     "/api/livestream/abc/report",
			cookie:   validCookie,
			usecase:  &fakeLivecommentUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "returns 403 when not the owner",
			path:     "/api/livestream/10/report",
			cookie:   validCookie,
			usecase:  &fakeLivecommentUsecase{err: usecase.ErrNotLivestreamOwner},
			wantCode: http.StatusForbidden,
			wantBody: `{"message":"can't get other streamer's livecomment reports"}` + "\n",
		},
		{
			name:     "returns 500 on unexpected error",
			path:     "/api/livestream/10/report",
			cookie:   validCookie,
			usecase:  &fakeLivecommentUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/livestream/:livestream_id/report", NewLivecommentHandler(tt.usecase).GetLivecommentReports)

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
			if tt.wantCode == http.StatusOK && (tt.usecase.gotUserID != 2 || tt.usecase.gotLivestreamID != 10) {
				t.Errorf("userID = %d, livestreamID = %d", tt.usecase.gotUserID, tt.usecase.gotLivestreamID)
			}
		})
	}
}
