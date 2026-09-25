package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/labstack/echo/v4"
)

type fakeTagUsecase struct {
	tags []*model.TagModel
	err  error
}

func (s *fakeTagUsecase) FindAll(ctx context.Context) ([]*model.TagModel, error) {
	return s.tags, s.err
}

func TestTagHandler_GetTags(t *testing.T) {
	tests := []struct {
		name     string
		usecase  *fakeTagUsecase
		wantCode int
		wantBody string
	}{
		{
			name: "returns tags",
			usecase: &fakeTagUsecase{tags: []*model.TagModel{
				{ID: 1, Name: "ライブ配信"},
				{ID: 2, Name: "ゲーム実況"},
			}},
			wantCode: http.StatusOK,
			wantBody: `{"tags":[{"id":1,"name":"ライブ配信"},{"id":2,"name":"ゲーム実況"}]}` + "\n",
		},
		{
			name:     "returns empty array when no tags",
			usecase:  &fakeTagUsecase{tags: nil},
			wantCode: http.StatusOK,
			wantBody: `{"tags":[]}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/tag", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			if err := NewTagHandler(tt.usecase).GetTags(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rec.Code != tt.wantCode {
				t.Errorf("code = %d, want %d", rec.Code, tt.wantCode)
			}
			if rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestTagHandler_GetTags_Error(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/tag", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := NewTagHandler(&fakeTagUsecase{err: errors.New("boom")}).GetTags(c)

	if he, ok := errors.AsType[*echo.HTTPError](err); !ok || he.Code != http.StatusInternalServerError {
		t.Fatalf("err = %v, want 500 HTTPError", err)
	}
}
