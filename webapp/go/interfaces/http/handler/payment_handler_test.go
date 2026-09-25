package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

type fakePaymentUsecase struct {
	totalTip int64
	err      error
}

func (u *fakePaymentUsecase) FindTotalTip(ctx context.Context) (int64, error) {
	return u.totalTip, u.err
}

func TestPaymentHandler_GetPaymentResult(t *testing.T) {
	tests := []struct {
		name     string
		usecase  *fakePaymentUsecase
		wantCode int
		wantBody string
	}{
		{
			// セッションが無くても返す (移行前と同じ)
			name:     "returns total tip without session",
			usecase:  &fakePaymentUsecase{totalTip: 1500},
			wantCode: http.StatusOK,
			wantBody: `{"total_tip":1500}` + "\n",
		},
		{
			name:     "returns 500 on unexpected error",
			usecase:  &fakePaymentUsecase{err: errors.New("boom")},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"message":"boom"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, newPaymentHandler(tt.usecase).GetPaymentResult, testRequest{method: http.MethodGet, route: "/api/payment", path: "/api/payment"})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
		})
	}
}
