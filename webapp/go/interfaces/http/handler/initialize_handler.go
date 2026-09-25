package handler

import (
	"net/http"

	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type InitializeResponse struct {
	Language string `json:"language"`
}

type initializeHandler struct {
	initializeUsecase usecase.InitializeUsecase
}

func newInitializeHandler(initializeUsecase usecase.InitializeUsecase) *initializeHandler {
	return &initializeHandler{initializeUsecase: initializeUsecase}
}

// 初期化
// POST /api/initialize
func (h *initializeHandler) Initialize(c echo.Context) error {
	if err := h.initializeUsecase.Initialize(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// レスポンスには影響しないが、移行前と同じくリクエストヘッダに付けている
	c.Request().Header.Add("Content-Type", "application/json;charset=utf-8")
	return c.JSON(http.StatusOK, InitializeResponse{
		Language: "golang",
	})
}
