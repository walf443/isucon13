package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/labstack/echo/v4"
)

// 一覧取得 API で指定できる limit の上限。
// フロントエンドは最大 100、ベンチマーカーは最大 50 を指定する。
const (
	maxLivestreamsLimit  domain.Limit = 100
	maxLivecommentsLimit domain.Limit = 100
	maxReactionsLimit    domain.Limit = 100
)

// parseLimitQueryParam はクエリパラメータ limit を 1 以上 max 以下の値として読み取る。
// 指定が無い場合は nil を返す。不正な値の場合は 400 の echo.HTTPError を返す。
func parseLimitQueryParam(c echo.Context, max domain.Limit) (*domain.Limit, error) {
	s := c.QueryParam("limit")
	if s == "" {
		return nil, nil
	}
	limit, err := domain.ParseLimit(s, max)
	if errors.Is(err, domain.ErrLimitOutOfRange) {
		return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("limit query parameter must be between 1 and %d", max))
	}
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "limit query parameter must be integer")
	}
	return &limit, nil
}
