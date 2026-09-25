package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeInitializeUsecase struct {
	err error
}

func (u *fakeInitializeUsecase) Initialize(ctx context.Context) error {
	return u.err
}

func TestInitializeHandler_Initialize(t *testing.T) {
	tests := []struct {
		name     string
		usecase  *fakeInitializeUsecase
		wantCode int
		wantBody string
	}{
		{
			// セッションは不要
			name:     "initializes",
			usecase:  &fakeInitializeUsecase{},
			wantCode: http.StatusOK,
			wantBody: `{"language":"golang"}` + "\n",
		},
		{
			name:     "returns 500 on failure",
			usecase:  &fakeInitializeUsecase{err: errors.New("failed to initialize: exit status 1")},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"message":"failed to initialize: exit status 1"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newTestEcho()
			e.POST("/api/initialize", NewInitializeHandler(tt.usecase).Initialize)

			req := httptest.NewRequest(http.MethodPost, "/api/initialize", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("code = %d, want %d (body: %s)", rec.Code, tt.wantCode, rec.Body.String())
			}
			if rec.Body.String() != tt.wantBody {
				t.Errorf("body = %s\nwant   %s", rec.Body.String(), tt.wantBody)
			}
		})
	}
}
