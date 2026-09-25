package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
)

type fakeLivecommentUsecase struct {
	livecomment  *model.Livecomment
	livecomments []*model.Livecomment
	report       *model.LivecommentReport
	reports      []*model.LivecommentReport
	err          error

	gotLivestreamID  model.LivestreamID
	gotLivecommentID model.LivecommentID
	gotLimit         *model.Limit
	gotUserID        model.UserID
	gotComment       string
	gotTip           int64
}

func (u *fakeLivecommentUsecase) Create(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID, comment string, tip int64) (*model.Livecomment, error) {
	u.gotUserID = userID
	u.gotLivestreamID = livestreamID
	u.gotComment = comment
	u.gotTip = tip
	return u.livecomment, u.err
}

func (u *fakeLivecommentUsecase) Report(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID, livecommentID model.LivecommentID) (*model.LivecommentReport, error) {
	u.gotUserID = userID
	u.gotLivestreamID = livestreamID
	u.gotLivecommentID = livecommentID
	return u.report, u.err
}

func (u *fakeLivecommentUsecase) FindAllByLivestreamID(ctx context.Context, livestreamID model.LivestreamID, limit *model.Limit) ([]*model.Livecomment, error) {
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
	tests := []struct {
		name      string
		path      string
		cookie    func(t *testing.T) *http.Cookie
		usecase   *fakeLivecommentUsecase
		wantCode  int
		wantBody  string
		wantLimit *model.Limit
	}{
		{
			name:     "returns livecomments",
			path:     "/api/livestream/10/livecomment",
			cookie:   sessionAs(1),
			usecase:  &fakeLivecommentUsecase{livecomments: []*model.Livecomment{testLivecomment}},
			wantCode: http.StatusOK,
			wantBody: "[" + testLivecommentJSON + "]\n",
		},
		{
			name:      "passes limit and returns empty array",
			path:      "/api/livestream/10/livecomment?limit=5",
			cookie:    sessionAs(1),
			usecase:   &fakeLivecommentUsecase{},
			wantCode:  http.StatusOK,
			wantBody:  "[]\n",
			wantLimit: func() *model.Limit { l := model.Limit(5); return &l }(),
		},
		{
			name:     "returns 400 when livestream_id is not integer",
			path:     "/api/livestream/abc/livecomment",
			cookie:   sessionAs(1),
			usecase:  &fakeLivecommentUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on unexpected error",
			path:     "/api/livestream/10/livecomment",
			cookie:   sessionAs(1),
			usecase:  &fakeLivecommentUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newLivecommentHandler(tt.usecase).GetLivecomments, testRequest{method: http.MethodGet, route: "/api/livestream/:livestream_id/livecomment", path: tt.path, cookie: tt.cookie})
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

func TestLivecommentHandler_GetLivecomments_Limit(t *testing.T) {
	testLimitQueryParam(t, maxLivecommentsLimit, func(t *testing.T, limit string) (*httptest.ResponseRecorder, *model.Limit) {
		u := &fakeLivecommentUsecase{}
		rec := serve(t, newLivecommentHandler(u).GetLivecomments, testRequest{method: http.MethodGet, route: "/api/livestream/:livestream_id/livecomment", path: "/api/livestream/10/livecomment?limit=" + limit, cookie: sessionAs(1)})
		return rec, u.gotLimit
	})
}

func TestLivecommentHandler_GetLivecommentReports(t *testing.T) {
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
			cookie:   sessionAs(2),
			usecase:  &fakeLivecommentUsecase{reports: []*model.LivecommentReport{report}},
			wantCode: http.StatusOK,
			wantBody: `[{"id":7,"reporter":{"id":3,"name":"carol","theme":{"id":13,"dark_mode":false},"icon_hash":"ccc"},"livecomment":` + testLivecommentJSON + `,"created_at":1700000100}]` + "\n",
		},
		{
			name:     "returns empty array when no reports",
			path:     "/api/livestream/10/report",
			cookie:   sessionAs(2),
			usecase:  &fakeLivecommentUsecase{},
			wantCode: http.StatusOK,
			wantBody: "[]\n",
		},
		{
			name:     "returns 400 when livestream_id is not integer",
			path:     "/api/livestream/abc/report",
			cookie:   sessionAs(2),
			usecase:  &fakeLivecommentUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "returns 403 when not the owner",
			path:     "/api/livestream/10/report",
			cookie:   sessionAs(2),
			usecase:  &fakeLivecommentUsecase{err: usecase.ErrNotLivestreamOwner},
			wantCode: http.StatusForbidden,
			wantBody: errorBody(http.StatusForbidden, "can't get other streamer's livecomment reports"),
		},
		{
			name:     "returns 500 on unexpected error",
			path:     "/api/livestream/10/report",
			cookie:   sessionAs(2),
			usecase:  &fakeLivecommentUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newLivecommentHandler(tt.usecase).GetLivecommentReports, testRequest{method: http.MethodGet, route: "/api/livestream/:livestream_id/report", path: tt.path, cookie: tt.cookie})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
			if tt.wantCode == http.StatusOK && (tt.usecase.gotUserID != 2 || tt.usecase.gotLivestreamID != 10) {
				t.Errorf("userID = %d, livestreamID = %d", tt.usecase.gotUserID, tt.usecase.gotLivestreamID)
			}
		})
	}
}

func TestLivecommentHandler_PostLivecomment(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		cookie   func(t *testing.T) *http.Cookie
		body     string
		usecase  *fakeLivecommentUsecase
		wantCode int
		wantBody string
	}{
		{
			name:     "creates livecomment",
			path:     "/api/livestream/10/livecomment",
			cookie:   sessionAs(1),
			body:     `{"comment":"hello","tip":100}`,
			usecase:  &fakeLivecommentUsecase{livecomment: testLivecomment},
			wantCode: http.StatusCreated,
			wantBody: testLivecommentJSON + "\n",
		},
		{
			name:     "returns 400 when livestream_id is not integer",
			path:     "/api/livestream/abc/livecomment",
			cookie:   sessionAs(1),
			body:     `{"comment":"hello","tip":100}`,
			usecase:  &fakeLivecommentUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "returns 400 on invalid json",
			path:     "/api/livestream/10/livecomment",
			cookie:   sessionAs(1),
			body:     `{`,
			usecase:  &fakeLivecommentUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "returns 404 when livestream is not found",
			path:     "/api/livestream/10/livecomment",
			cookie:   sessionAs(1),
			body:     `{"comment":"hello","tip":100}`,
			usecase:  &fakeLivecommentUsecase{err: usecase.ErrLivestreamNotFound},
			wantCode: http.StatusNotFound,
			wantBody: errorBody(http.StatusNotFound, "livestream not found"),
		},
		{
			name:     "returns 400 when judged as spam",
			path:     "/api/livestream/10/livecomment",
			cookie:   sessionAs(1),
			body:     `{"comment":"bad","tip":0}`,
			usecase:  &fakeLivecommentUsecase{err: usecase.ErrSpamLivecomment},
			wantCode: http.StatusBadRequest,
			wantBody: errorBody(http.StatusBadRequest, "このコメントがスパム判定されました"),
		},
		{
			name:     "returns 500 on unexpected error",
			path:     "/api/livestream/10/livecomment",
			cookie:   sessionAs(1),
			body:     `{"comment":"hello","tip":100}`,
			usecase:  &fakeLivecommentUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newLivecommentHandler(tt.usecase).PostLivecomment, testRequest{method: http.MethodPost, route: "/api/livestream/:livestream_id/livecomment", path: tt.path, body: tt.body, cookie: tt.cookie})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
			if tt.wantCode == http.StatusCreated {
				if tt.usecase.gotUserID != 1 || tt.usecase.gotLivestreamID != 10 || tt.usecase.gotComment != "hello" || tt.usecase.gotTip != 100 {
					t.Errorf("userID = %d, livestreamID = %d, comment = %q, tip = %d", tt.usecase.gotUserID, tt.usecase.gotLivestreamID, tt.usecase.gotComment, tt.usecase.gotTip)
				}
			}
		})
	}
}

func TestLivecommentHandler_PostLivecommentReport(t *testing.T) {
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
			name:     "creates report",
			path:     "/api/livestream/10/livecomment/50/report",
			cookie:   sessionAs(3),
			usecase:  &fakeLivecommentUsecase{report: report},
			wantCode: http.StatusCreated,
			wantBody: `{"id":7,"reporter":{"id":3,"name":"carol","theme":{"id":13,"dark_mode":false},"icon_hash":"ccc"},"livecomment":` + testLivecommentJSON + `,"created_at":1700000100}` + "\n",
		},
		{
			name:     "returns 400 when livestream_id is not integer",
			path:     "/api/livestream/abc/livecomment/50/report",
			cookie:   sessionAs(3),
			usecase:  &fakeLivecommentUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: errorBody(http.StatusBadRequest, "livestream_id in path must be integer"),
		},
		{
			name:     "returns 400 when livecomment_id is not integer",
			path:     "/api/livestream/10/livecomment/abc/report",
			cookie:   sessionAs(3),
			usecase:  &fakeLivecommentUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: errorBody(http.StatusBadRequest, "livecomment_id in path must be integer"),
		},
		{
			name:     "returns 404 when livestream is not found",
			path:     "/api/livestream/10/livecomment/50/report",
			cookie:   sessionAs(3),
			usecase:  &fakeLivecommentUsecase{err: usecase.ErrLivestreamNotFound},
			wantCode: http.StatusNotFound,
			wantBody: errorBody(http.StatusNotFound, "livestream not found"),
		},
		{
			name:     "returns 404 when livecomment is not found",
			path:     "/api/livestream/10/livecomment/50/report",
			cookie:   sessionAs(3),
			usecase:  &fakeLivecommentUsecase{err: usecase.ErrLivecommentNotFound},
			wantCode: http.StatusNotFound,
			wantBody: errorBody(http.StatusNotFound, "livecomment not found"),
		},
		{
			name:     "returns 500 on unexpected error",
			path:     "/api/livestream/10/livecomment/50/report",
			cookie:   sessionAs(3),
			usecase:  &fakeLivecommentUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newLivecommentHandler(tt.usecase).PostLivecommentReport, testRequest{method: http.MethodPost, route: "/api/livestream/:livestream_id/livecomment/:livecomment_id/report", path: tt.path, cookie: tt.cookie})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
			if tt.wantCode == http.StatusCreated && (tt.usecase.gotUserID != 3 || tt.usecase.gotLivestreamID != 10 || tt.usecase.gotLivecommentID != 50) {
				t.Errorf("userID = %d, livestreamID = %d, livecommentID = %d", tt.usecase.gotUserID, tt.usecase.gotLivestreamID, tt.usecase.gotLivecommentID)
			}
		})
	}
}
