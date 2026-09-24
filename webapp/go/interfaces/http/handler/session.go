package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const (
	DefaultSessionIDKey      = "SESSIONID"
	DefaultSessionExpiresKey = "EXPIRES"
	DefaultUserIDKey         = "USERID"
	DefaultUsernameKey       = "USERNAME"
)

func VerifyUserSession(c echo.Context) error {
	sess, err := session.Get(DefaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to get session")
	}

	sessionExpires, ok := sess.Values[DefaultSessionExpiresKey]
	if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "failed to get EXPIRES value from session")
	}

	_, ok = sess.Values[DefaultUserIDKey].(int64)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to get USERID value from session")
	}

	now := time.Now()
	if now.Unix() > sessionExpires.(int64) {
		return echo.NewHTTPError(http.StatusUnauthorized, "session has expired")
	}

	return nil
}

// getSessionUserID はセッションからログイン中のユーザIDを取り出す。
// VerifyUserSession で検証済みであることを前提とする。
func getSessionUserID(c echo.Context) (int64, error) {
	sess, err := session.Get(DefaultSessionIDKey, c)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusUnauthorized, "failed to get session")
	}
	userID, ok := sess.Values[DefaultUserIDKey].(int64)
	if !ok {
		return 0, echo.NewHTTPError(http.StatusUnauthorized, "failed to get USERID value from session")
	}
	return userID, nil
}
