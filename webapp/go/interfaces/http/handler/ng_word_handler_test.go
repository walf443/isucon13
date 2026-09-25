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

type fakeNGWordUsecase struct {
	ngWords []*model.NGWordModel
	err     error

	gotUserID       model.UserID
	gotLivestreamID model.LivestreamID
}

func (u *fakeNGWordUsecase) FindAllByLivestreamID(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID) ([]*model.NGWordModel, error) {
	u.gotUserID = userID
	u.gotLivestreamID = livestreamID
	return u.ngWords, u.err
}

func TestNGWordHandler_GetNGWords(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 2, time.Now().Add(time.Hour))
	}

	tests := []struct {
		name     string
		path     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeNGWordUsecase
		wantCode int
		wantBody string
	}{
		{
			name:   "returns NG words",
			path:   "/api/livestream/10/ngwords",
			cookie: validCookie,
			usecase: &fakeNGWordUsecase{ngWords: []*model.NGWordModel{
				{ID: 2, UserID: 2, LivestreamID: 10, Word: "bad2", CreatedAt: 200},
				{ID: 1, UserID: 2, LivestreamID: 10, Word: "bad1", CreatedAt: 100},
			}},
			wantCode: http.StatusOK,
			wantBody: `[{"id":2,"user_id":2,"livestream_id":10,"word":"bad2","created_at":200},{"id":1,"user_id":2,"livestream_id":10,"word":"bad1","created_at":100}]` + "\n",
		},
		{
			// NG ワードが無い場合は [] ではなく null を返す (移行前と同じ)
			name:     "returns null when no NG words",
			path:     "/api/livestream/10/ngwords",
			cookie:   validCookie,
			usecase:  &fakeNGWordUsecase{ngWords: nil},
			wantCode: http.StatusOK,
			wantBody: "null\n",
		},
		{
			name:     "returns 403 without session",
			path:     "/api/livestream/10/ngwords",
			cookie:   nil,
			usecase:  &fakeNGWordUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 400 when livestream_id is not integer",
			path:     "/api/livestream/abc/ngwords",
			cookie:   validCookie,
			usecase:  &fakeNGWordUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "returns 500 on unexpected error",
			path:     "/api/livestream/10/ngwords",
			cookie:   validCookie,
			usecase:  &fakeNGWordUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/livestream/:livestream_id/ngwords", NewNGWordHandler(tt.usecase).GetNGWords)

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
