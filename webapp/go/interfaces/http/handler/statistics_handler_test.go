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

type fakeStatisticsUsecase struct {
	userStatistics       *model.UserStatistics
	livestreamStatistics *model.LivestreamStatistics
	err                  error

	gotUsername     string
	gotLivestreamID model.LivestreamID
}

func (u *fakeStatisticsUsecase) FindLivestreamStatistics(ctx context.Context, livestreamID model.LivestreamID) (*model.LivestreamStatistics, error) {
	u.gotLivestreamID = livestreamID
	return u.livestreamStatistics, u.err
}

func (u *fakeStatisticsUsecase) FindUserStatistics(ctx context.Context, username string) (*model.UserStatistics, error) {
	u.gotUsername = username
	return u.userStatistics, u.err
}

func TestStatisticsHandler_GetUserStatistics(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 1, time.Now().Add(time.Hour))
	}

	tests := []struct {
		name     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeStatisticsUsecase
		wantCode int
		wantBody string
	}{
		{
			name:   "returns statistics",
			cookie: validCookie,
			usecase: &fakeStatisticsUsecase{userStatistics: &model.UserStatistics{
				Rank:              2,
				ViewersCount:      7,
				TotalReactions:    5,
				TotalLivecomments: 3,
				TotalTip:          25,
				FavoriteEmoji:     "smile",
			}},
			wantCode: http.StatusOK,
			wantBody: `{"rank":2,"viewers_count":7,"total_reactions":5,"total_livecomments":3,"total_tip":25,"favorite_emoji":"smile"}` + "\n",
		},
		{
			name:     "returns 403 without session",
			cookie:   nil,
			usecase:  &fakeStatisticsUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			// 他のエンドポイントと違い 404 ではなく 400 (移行前と同じ)
			name:     "returns 400 when user is not found",
			cookie:   validCookie,
			usecase:  &fakeStatisticsUsecase{err: usecase.ErrUserNotFound},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"not found user that has the given username"}` + "\n",
		},
		{
			name:     "returns 500 on unexpected error",
			cookie:   validCookie,
			usecase:  &fakeStatisticsUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"message":"boom"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/user/:username/statistics", newStatisticsHandler(tt.usecase).GetUserStatistics)

			req := httptest.NewRequest(http.MethodGet, "/api/user/bob/statistics", nil)
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
			if tt.wantCode == http.StatusOK && tt.usecase.gotUsername != "bob" {
				t.Errorf("username = %q, want bob", tt.usecase.gotUsername)
			}
		})
	}
}

func TestStatisticsHandler_GetLivestreamStatistics(t *testing.T) {
	validCookie := func(t *testing.T) *http.Cookie {
		return newSessionCookie(t, 1, time.Now().Add(time.Hour))
	}

	tests := []struct {
		name     string
		path     string
		cookie   func(t *testing.T) *http.Cookie
		usecase  *fakeStatisticsUsecase
		wantCode int
		wantBody string
	}{
		{
			name:   "returns statistics",
			path:   "/api/livestream/10/statistics",
			cookie: validCookie,
			usecase: &fakeStatisticsUsecase{livestreamStatistics: &model.LivestreamStatistics{
				Rank:           2,
				ViewersCount:   7,
				TotalReactions: 5,
				TotalReports:   1,
				MaxTip:         300,
			}},
			wantCode: http.StatusOK,
			wantBody: `{"rank":2,"viewers_count":7,"total_reactions":5,"total_reports":1,"max_tip":300}` + "\n",
		},
		{
			name:     "returns 403 without session",
			path:     "/api/livestream/10/statistics",
			cookie:   nil,
			usecase:  &fakeStatisticsUsecase{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "returns 400 when livestream_id is not integer",
			path:     "/api/livestream/abc/statistics",
			cookie:   validCookie,
			usecase:  &fakeStatisticsUsecase{},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"livestream_id in path must be integer"}` + "\n",
		},
		{
			// 404 ではなく 400 (移行前と同じ)
			name:     "returns 400 when livestream is not found",
			path:     "/api/livestream/10/statistics",
			cookie:   validCookie,
			usecase:  &fakeStatisticsUsecase{err: usecase.ErrLivestreamNotFound},
			wantCode: http.StatusBadRequest,
			wantBody: `{"message":"cannot get stats of not found livestream"}` + "\n",
		},
		{
			name:     "returns 500 on unexpected error",
			path:     "/api/livestream/10/statistics",
			cookie:   validCookie,
			usecase:  &fakeStatisticsUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"message":"boom"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.GET("/api/livestream/:livestream_id/statistics", newStatisticsHandler(tt.usecase).GetLivestreamStatistics)

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
			if tt.wantCode == http.StatusOK && tt.usecase.gotLivestreamID != 10 {
				t.Errorf("livestreamID = %d, want 10", tt.usecase.gotLivestreamID)
			}
		})
	}
}
