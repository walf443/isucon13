package handler

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// Setup は e にセッション・ルーティング・エラーレスポンスを設定する。
// fallbackImagePath はアイコン未登録のユーザに返す画像ファイルのパス。
func Setup(e *echo.Echo, sessionSecret []byte, u Usecases, fallbackImagePath string) {
	e.Use(session.Middleware(newSessionStore(sessionSecret)))
	RegisterRoutes(e, u, fallbackImagePath)
	e.HTTPErrorHandler = errorResponseHandler
}

// newSessionStore はセッションを保存する Cookie のストアを返す。
func newSessionStore(secret []byte) *sessions.CookieStore {
	cookieStore := sessions.NewCookieStore(secret)
	cookieStore.Options.Domain = "*.u.isucon.dev"
	return cookieStore
}

type errorResponse struct {
	Error string `json:"error"`
}

// errorResponseHandler は handler が返したエラーを {"error": "..."} の形でレスポンスにする。
// echo.HTTPError の場合、本文は "code=<ステータスコード>, message=<メッセージ>" になる (移行前と同じ)。
func errorResponseHandler(err error, c echo.Context) {
	c.Logger().Errorf("error at %s: %+v", c.Path(), err)
	if he, ok := err.(*echo.HTTPError); ok {
		if e := c.JSON(he.Code, &errorResponse{Error: err.Error()}); e != nil {
			c.Logger().Errorf("%+v", e)
		}
		return
	}

	if e := c.JSON(http.StatusInternalServerError, &errorResponse{Error: err.Error()}); e != nil {
		c.Logger().Errorf("%+v", e)
	}
}
