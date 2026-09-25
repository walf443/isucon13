package handler

import (
	"net/http"

	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type InitializeResponse struct {
	Language string `json:"language"`
}

type InitializeHandler struct {
	initializeUsecase usecase.InitializeUsecase
}

func NewInitializeHandler(initializeUsecase usecase.InitializeUsecase) *InitializeHandler {
	return &InitializeHandler{initializeUsecase: initializeUsecase}
}

// 初期化
// POST /api/initialize
func (h *InitializeHandler) Initialize(c echo.Context) error {
	if err := h.initializeUsecase.Initialize(c.Request().Context()); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// レスポンスには影響しないが、移行前と同じくリクエストヘッダに付けている
	c.Request().Header.Add("Content-Type", "application/json;charset=utf-8")
	return c.JSON(http.StatusOK, InitializeResponse{
		Language: "golang",
	})
}
