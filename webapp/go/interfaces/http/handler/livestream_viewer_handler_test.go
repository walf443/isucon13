package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
)

type fakeLivestreamViewerUsecase struct {
	err error

	called          string
	gotUserID       model.UserID
	gotLivestreamID model.LivestreamID
}

func (u *fakeLivestreamViewerUsecase) Enter(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) error {
	u.called = "Enter"
	u.gotUserID = userID
	u.gotLivestreamID = livestreamID
	return u.err
}

func (u *fakeLivestreamViewerUsecase) Exit(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) error {
	u.called = "Exit"
	u.gotUserID = userID
	u.gotLivestreamID = livestreamID
	return u.err
}

func TestLivestreamViewerHandler(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 1, time.Now().Add(time.Hour))
	}

	tests := []struct {
		name       string
		method     string
		path       string
		cookie     func(t *testing.T) *http.Cookie
		usecase    *fakeLivestreamViewerUsecase
		wantCode   int
		wantBody   string
		wantCalled string
	}{
		{
			name:       "enter",
			method:     http.MethodPost,
			path:       "/api/livestream/10/enter",
			cookie:     validCookie,
			usecase:    &fakeLivestreamViewerUsecase{},
			wantCode:   http.StatusOK,
			wantCalled: "Enter",
		},
		{
			name:     "enter returns 403 without session",
			method:   http.MethodPost,
			path:     "/api/livestream/10/enter",
			cookie:   nil,
			usecase:  &fakeLivestreamViewerUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			// enter だけメッセージに "in path" が無い (移行前と同じ)
			name:     "enter returns 400 when livestream_id is not integer",
			method:   http.MethodPost,
			path:     "/api/livestream/abc/enter",
			cookie:   validCookie,
			usecase:  &fakeLivestreamViewerUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"livestream_id must be integer"}` + "\n",
		},
		{
			name:       "enter returns 500 on unexpected error",
			method:     http.MethodPost,
			path:       "/api/livestream/10/enter",
			cookie:     validCookie,
			usecase:    &fakeLivestreamViewerUsecase{err: errors.New("boom")},
			wantCode:   http.StatusInternalServerError,
			wantBody:   `{"message":"boom"}` + "\n",
			wantCalled: "Enter",
		},
		{
			name:       "exit",
			method:     http.MethodDelete,
			path:       "/api/livestream/10/exit",
			cookie:     validCookie,
			usecase:    &fakeLivestreamViewerUsecase{},
			wantCode:   http.StatusOK,
			wantCalled: "Exit",
		},
		{
			name:     "exit returns 403 without session",
			method:   http.MethodDelete,
			path:     "/api/livestream/10/exit",
			cookie:   nil,
			usecase:  &fakeLivestreamViewerUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "exit returns 400 when livestream_id is not integer",
			method:   http.MethodDelete,
			path:     "/api/livestream/abc/exit",
			cookie:   validCookie,
			usecase:  &fakeLivestreamViewerUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"livestream_id in path must be integer"}` + "\n",
		},
		{
			name:       "exit returns 500 on unexpected error",
			method:     http.MethodDelete,
			path:       "/api/livestream/10/exit",
			cookie:     validCookie,
			usecase:    &fakeLivestreamViewerUsecase{err: errors.New("boom")},
			wantCode:   http.StatusInternalServerError,
			wantBody:   `{"message":"boom"}` + "\n",
			wantCalled: "Exit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			h := NewLivestreamViewerHandler(tt.usecase)
			e.POST("/api/livestream/:livestream_id/enter", h.EnterLivestream)
			e.DELETE("/api/livestream/:livestream_id/exit", h.ExitLivestream)

			req := httptest.NewRequest(tt.method, tt.path, nil)
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
			if tt.usecase.called != tt.wantCalled {
				t.Errorf("called = %q, want %q", tt.usecase.called, tt.wantCalled)
			}
			if tt.wantCalled != "" && (tt.usecase.gotUserID != 1 || tt.usecase.gotLivestreamID != 10) {
				t.Errorf("userID = %d, livestreamID = %d, want 1, 10", tt.usecase.gotUserID, tt.usecase.gotLivestreamID)
			}
		})
	}
}
