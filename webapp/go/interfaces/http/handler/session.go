package handler

import (
	"net/http"
	"time"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const (
	defaultSessionIDKey      = "SESSIONID"
	defaultSessionExpiresKey = "EXPIRES"
	defaultUserIDKey         = "USERID"
	defaultUsernameKey       = "USERNAME"
)

// requireSession はログイン済みの有効なセッションを必須にする middleware を返す。
// now は現在時刻を返す。テストで差し替えられるようにしている。
func requireSession(now func() time.Time) echo.MiddlewareFunc {
	return newSessionMiddleware(now, nil)
}

// requireSessionWithLog は requireSession と同じだが、検証に失敗したときにログを出す。
// 移行前から GET /api/user/:username/theme だけがこのログを出している。
func requireSessionWithLog(now func() time.Time) echo.MiddlewareFunc {
	return newSessionMiddleware(now, func(c echo.Context, err error) {
		c.Logger().Printf("verifyUserSession: %+v\n", err)
	})
}

// newSessionMiddleware はセッションを検証する middleware を返す。onError が nil でなければ検証の失敗時に呼ぶ。
func newSessionMiddleware(now func() time.Time, onError func(c echo.Context, err error)) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if err := verifyUserSession(c, now()); err != nil {
				if onError != nil {
					onError(c, err)
				}
				// echo.NewHTTPErrorが返っているのでそのまま出力
				return err
			}
			return next(c)
		}
	}
}

// verifyUserSession はセッションがログイン済みで、now の時点で有効期限内であることを確認する。
func verifyUserSession(c echo.Context, now time.Time) error {
	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to get session")
	}

	sessionExpires, ok := sess.Values[defaultSessionExpiresKey]
	if !ok {
		return echo.NewHTTPError(http.StatusForbidden, "failed to get EXPIRES value from session")
	}

	_, ok = sess.Values[defaultUserIDKey].(int64)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "failed to get USERID value from session")
	}

	if now.Unix() > sessionExpires.(int64) {
		return echo.NewHTTPError(http.StatusUnauthorized, "session has expired")
	}

	return nil
}

// getSessionUserID はセッションからログイン中のユーザIDを取り出す。
// requireSession で検証済みであることを前提とする。
func getSessionUserID(c echo.Context) (model.UserID, error) {
	sess, err := session.Get(defaultSessionIDKey, c)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusUnauthorized, "failed to get session")
	}
	userID, ok := sess.Values[defaultUserIDKey].(int64)
	if !ok {
		return 0, echo.NewHTTPError(http.StatusUnauthorized, "failed to get USERID value from session")
	}
	return model.UserID(userID), nil
}
