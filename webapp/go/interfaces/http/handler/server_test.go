package handler

import (
	"errors"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
)

// Setup で組み立てた echo (本番と同じ構成) で、セッションの検証とエラーレスポンスが効いていることを確認する。
func TestSetup(t *testing.T) {
	e := echo.New()
	Setup(e, []byte("test-secret"), Usecases{}, "")

	rec := send(t, e, testRequest{method: http.MethodGet, path: "/api/user/me"})
	assertResponse(t, rec, http.StatusForbidden, `{"error":"code=403, message=failed to get EXPIRES value from session"}`+"\n")
}

func TestErrorResponseHandler(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantBody string
	}{
		{
			name:     "HTTPError",
			err:      echo.NewHTTPError(http.StatusNotFound, "not found"),
			wantCode: http.StatusNotFound,
			wantBody: `{"error":"code=404, message=not found"}` + "\n",
		},
		{
			// echo.HTTPError 以外は 500 で、エラーメッセージをそのまま返す (移行前と同じ)
			name:     "other error",
			err:      errors.New("boom"),
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"boom"}` + "\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, func(c echo.Context) error { return tt.err }, testRequest{method: http.MethodGet, route: "/", path: "/"})
			assertResponse(t, rec, tt.wantCode, tt.wantBody)
		})
	}
}
