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
	"github.com/isucon/isucon13/webapp/go/usecase"
)

type fakeNGWordUsecase struct {
	ngWords []*model.NGWordModel
	wordID  model.NGWordID
	err     error
	gotWord string

	gotUserID       model.UserID
	gotLivestreamID model.LivestreamID
}

func (u *fakeNGWordUsecase) Moderate(ctx context.Context, userID model.UserID, livestreamID model.LivestreamID, word string) (model.NGWordID, error) {
	u.gotUserID = userID
	u.gotLivestreamID = livestreamID
	u.gotWord = word
	return u.wordID, u.err
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
			e.GET("/api/livestream/:livestream_id/ngwords", newNGWordHandler(tt.usecase).GetNGWords)

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

func TestNGWordHandler_Moderate(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 2, time.Now().Add(time.Hour))
	}

	tests := []struct {
		name     string
		path     string
		cookie   func(t *testing.T) *http.Cookie
		body     string
		usecase  *fakeNGWordUsecase
		wantCode int
		wantBody string
	}{
		{
			name:     "registers NG word",
			path:     "/api/livestream/10/moderate",
			cookie:   validCookie,
			body:     `{"ng_word":"bad"}`,
			usecase:  &fakeNGWordUsecase{wordID: 7},
			wantCode: http.StatusCreated,
			wantBody: `{"word_id":7}` + "\n",
		},
		{
			name:     "returns 403 without session",
			path:     "/api/livestream/10/moderate",
			cookie:   nil,
			body:     `{"ng_word":"bad"}`,
			usecase:  &fakeNGWordUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 400 when livestream_id is not integer",
			path:     "/api/livestream/abc/moderate",
			cookie:   validCookie,
			body:     `{"ng_word":"bad"}`,
			usecase:  &fakeNGWordUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "returns 400 on invalid json",
			path:     "/api/livestream/10/moderate",
			cookie:   validCookie,
			body:     `{`,
			usecase:  &fakeNGWordUsecase{},
			wantCode: http.StatusBadRequest,
		},
		{
			// 他の配信者のライブ配信は 403 ではなく 400 (移行前と同じ)
			name:     "returns 400 when not the owner",
			path:     "/api/livestream/10/moderate",
			cookie:   validCookie,
			body:     `{"ng_word":"bad"}`,
			usecase:  &fakeNGWordUsecase{err: usecase.ErrNotLivestreamOwner},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"A streamer can't moderate livestreams that other streamers own"}` + "\n",
		},
		{
			name:     "returns 500 on unexpected error",
			path:     "/api/livestream/10/moderate",
			cookie:   validCookie,
			body:     `{"ng_word":"bad"}`,
			usecase:  &fakeNGWordUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.POST("/api/livestream/:livestream_id/moderate", newNGWordHandler(tt.usecase).Moderate)

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
			if tt.wantCode == http.StatusCreated && (tt.usecase.gotUserID != 2 || tt.usecase.gotLivestreamID != 10 || tt.usecase.gotWord != "bad") {
				t.Errorf("userID = %d, livestreamID = %d, word = %q", tt.usecase.gotUserID, tt.usecase.gotLivestreamID, tt.usecase.gotWord)
			}
		})
	}
}
