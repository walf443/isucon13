package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type Reaction struct {
	ID         model.ReactionID `json:"id"`
	EmojiName  string           `json:"emoji_name"`
	User       User             `json:"user"`
	Livestream Livestream       `json:"livestream"`
	CreatedAt  int64            `json:"created_at"`
}

func newReaction(r *model.Reaction) Reaction {
	return Reaction{
		ID:         r.ID,
		EmojiName:  r.EmojiName,
		User:       newUser(&r.User),
		Livestream: newLivestream(&r.Livestream),
		CreatedAt:  r.CreatedAt,
	}
}

type PostReactionRequest struct {
	EmojiName string `json:"emoji_name"`
}

type ReactionHandler struct {
	reactionUsecase usecase.ReactionUsecase
}

func NewReactionHandler(reactionUsecase usecase.ReactionUsecase) *ReactionHandler {
	return &ReactionHandler{reactionUsecase: reactionUsecase}
}

// GET /api/livestream/:livestream_id/reaction
func (h *ReactionHandler) GetReactions(c echo.Context) error {
	ctx := c.Request().Context()

	if err := VerifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	livestreamID, err := model.ParseLivestreamID(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	var limit *int64
	if c.QueryParam("limit") != "" {
		l, err := strconv.Atoi(c.QueryParam("limit"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "limit query parameter must be integer")
		}
		limit64 := int64(l)
		limit = &limit64
	}

	reactionModels, err := h.reactionUsecase.FindAllByLivestreamID(ctx, livestreamID, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	reactions := make([]Reaction, len(reactionModels))
	for i, r := range reactionModels {
		reactions[i] = newReaction(r)
	}
	return c.JSON(http.StatusOK, reactions)
}

// POST /api/livestream/:livestream_id/reaction
func (h *ReactionHandler) PostReaction(c echo.Context) error {
	ctx := c.Request().Context()
	livestreamID, err := model.ParseLivestreamID(c.Param("livestream_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "livestream_id in path must be integer")
	}

	if err := VerifyUserSession(c); err != nil {
		// echo.NewHTTPErrorが返っているのでそのまま出力
		return err
	}

	userID, err := getSessionUserID(c)
	if err != nil {
		return err
	}

	var req PostReactionRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode the request body as json")
	}

	reaction, err := h.reactionUsecase.Create(ctx, userID, livestreamID, req.EmojiName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, newReaction(reaction))
}
