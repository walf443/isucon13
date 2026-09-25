package handler

import (
	"net/http"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/labstack/echo/v4"
)

type Tag struct {
	ID   model.TagID `json:"id"`
	Name string      `json:"name"`
}

type TagsResponse struct {
	Tags []*Tag `json:"tags"`
}

type tagHandler struct {
	tagUsecase usecase.TagUsecase
}

func newTagHandler(tagUsecase usecase.TagUsecase) *tagHandler {
	return &tagHandler{tagUsecase: tagUsecase}
}

// GET /api/tag
func (h *tagHandler) GetTags(c echo.Context) error {
	ctx := c.Request().Context()

	tagModels, err := h.tagUsecase.FindAll(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	tags := make([]*Tag, len(tagModels))
	for i := range tagModels {
		tags[i] = &Tag{
			ID:   tagModels[i].ID,
			Name: tagModels[i].Name,
		}
	}
	return c.JSON(http.StatusOK, &TagsResponse{
		Tags: tags,
	})
}
