package handler

import (
	"net/http"

	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type paymentResultResponse struct {
	TotalTip int64 `json:"total_tip"`
}

type paymentHandler struct {
	paymentUsecase usecase.PaymentUsecase
}

func newPaymentHandler(paymentUsecase usecase.PaymentUsecase) *paymentHandler {
	return &paymentHandler{paymentUsecase: paymentUsecase}
}

// 課金情報
// GET /api/payment
func (h *paymentHandler) GetPaymentResult(c echo.Context) error {
	ctx := c.Request().Context()

	// 移行前と同じくセッションは確認しない
	totalTip, err := h.paymentUsecase.FindTotalTip(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, &paymentResultResponse{
		TotalTip: totalTip,
	})
}
