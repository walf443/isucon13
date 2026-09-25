package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain"
)

type fakeLivestreamViewerUsecase struct {
	err error

	called          string
	gotUserID       domain.UserID
	gotLivestreamID domain.LivestreamID
}

func (u *fakeLivestreamViewerUsecase) Enter(ctx context.Context, userID domain.UserID, livestreamID domain.LivestreamID) error {
	u.called = "Enter"
	u.gotUserID = userID
	u.gotLivestreamID = livestreamID
	return u.err
}

func (u *fakeLivestreamViewerUsecase) Exit(ctx context.Context, userID domain.UserID, livestreamID domain.LivestreamID) error {
	u.called = "Exit"
	u.gotUserID = userID
	u.gotLivestreamID = livestreamID
	return u.err
}

func TestLivestreamViewerHandler(t *testing.T) {
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
			cookie:     sessionAs(1),
			usecase:    &fakeLivestreamViewerUsecase{},
			wantCode:   http.StatusOK,
			wantCalled: "Enter",
		},
		{
			// enter だけメッセージに "in path" が無い (移行前と同じ)
			name:     "enter returns 400 when livestream_id is not integer",
			method:   http.MethodPost,
			path:     "/api/livestream/abc/enter",
			cookie:   sessionAs(1),
			usecase:  &fakeLivestreamViewerUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: errorBody(http.StatusBadRequest, "livestream_id must be integer"),
		},
		{
			name:       "enter returns 500 on unexpected error",
			method:     http.MethodPost,
			path:       "/api/livestream/10/enter",
			cookie:     sessionAs(1),
			usecase:    &fakeLivestreamViewerUsecase{err: errors.New("boom")},
			wantCode:   http.StatusInternalServerError,
			wantBody:   errorBody(http.StatusInternalServerError, "boom"),
			wantCalled: "Enter",
		},
		{
			name:       "exit",
			method:     http.MethodDelete,
			path:       "/api/livestream/10/exit",
			cookie:     sessionAs(1),
			usecase:    &fakeLivestreamViewerUsecase{},
			wantCode:   http.StatusOK,
			wantCalled: "Exit",
		},
		{
			name:     "exit returns 400 when livestream_id is not integer",
			method:   http.MethodDelete,
			path:     "/api/livestream/abc/exit",
			cookie:   sessionAs(1),
			usecase:  &fakeLivestreamViewerUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: errorBody(http.StatusBadRequest, "livestream_id in path must be integer"),
		},
		{
			name:       "exit returns 500 on unexpected error",
			method:     http.MethodDelete,
			path:       "/api/livestream/10/exit",
			cookie:     sessionAs(1),
			usecase:    &fakeLivestreamViewerUsecase{err: errors.New("boom")},
			wantCode:   http.StatusInternalServerError,
			wantBody:   errorBody(http.StatusInternalServerError, "boom"),
			wantCalled: "Exit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newLivestreamViewerHandler(tt.usecase)
			route, handle := "/api/livestream/:livestream_id/enter", h.EnterLivestream
			if tt.method == http.MethodDelete {
				route, handle = "/api/livestream/:livestream_id/exit", h.ExitLivestream
			}
			rec := serve(t, handle, testRequest{method: tt.method, route: route, path: tt.path, cookie: tt.cookie})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
			if tt.usecase.called != tt.wantCalled {
				t.Errorf("called = %q, want %q", tt.usecase.called, tt.wantCalled)
			}
			if tt.wantCalled != "" && (tt.usecase.gotUserID != 1 || tt.usecase.gotLivestreamID != 10) {
				t.Errorf("userID = %d, livestreamID = %d, want 1, 10", tt.usecase.gotUserID, tt.usecase.gotLivestreamID)
			}
		})
	}
}
