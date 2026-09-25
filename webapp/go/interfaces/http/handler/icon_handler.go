package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type postIconRequest struct {
	Image []byte `json:"image"`
}

type postIconResponse struct {
	ID model.IconID `json:"id"`
}

type iconHandler struct {
	iconUsecase usecase.IconUsecase
	// fallbackImagePath はアイコン未登録のユーザに返す画像ファイルのパス。
	fallbackImagePath string
}

func newIconHandler(iconUsecase usecase.IconUsecase, fallbackImagePath string) *iconHandler {
	return &iconHandler{iconUsecase: iconUsecase, fallbackImagePath: fallbackImagePath}
}

// GET /api/user/:username/icon
func (h *iconHandler) GetIcon(c echo.Context) error {
	ctx := c.Request().Context()

	username := c.Param("username")

	image, err := h.iconUsecase.FindImageByUsername(ctx, username)
	if errors.Is(err, usecase.ErrUserNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "not found user that has the given username")
	}
	if errors.Is(err, usecase.ErrIconNotFound) {
		return c.File(h.fallbackImagePath)
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Blob(http.StatusOK, "image/jpeg", image)
}

// POST /api/icon
func (h *iconHandler) PostIcon(c echo.Context) error {
	ctx := c.Request().Context()

	if err := verifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	var req postIconRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	iconID, err := h.iconUsecase.Update(ctx, userID, req.Image)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, &postIconResponse{
		ID: iconID,
	})
}
