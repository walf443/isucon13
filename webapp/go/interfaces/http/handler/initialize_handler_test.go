package handler

import (
	"context"
	"errors"
	"net/http"
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
			wantBody: errorBody(http.StatusInternalServerError, "failed to initialize: exit status 1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newInitializeHandler(tt.usecase).Initialize, testRequest{method: http.MethodPost, route: "/api/initialize", path: "/api/initialize"})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
		})
	}
}
