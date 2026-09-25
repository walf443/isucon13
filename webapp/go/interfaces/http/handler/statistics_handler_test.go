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
	userStatistics *model.UserStatistics
	err            error

	gotUsername string
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
			e.GET("/api/user/:username/statistics", NewStatisticsHandler(tt.usecase).GetUserStatistics)

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
