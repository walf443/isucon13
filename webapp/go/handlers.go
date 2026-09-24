package main

import (
	"fmt"
	"os"

	infra "github.com/isucon/isucon13/webapp/go/infra/mysql"
	"github.com/isucon/isucon13/webapp/go/interfaces/http/handler"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/jmoiron/sqlx"
)

// handlers はクリーンアーキテクチャへ移行済みの handler をまとめたもの。
type handlers struct {
	tag   *handler.TagHandler
	theme *handler.ThemeHandler
	user  *handler.UserHandler
	icon  *handler.IconHandler
}

// newHandlers は repository・usecase・handler を組み立てる。
func newHandlers(db *sqlx.DB, fallbackImagePath string) (*handlers, error) {
	fallbackIcon, err := os.ReadFile(fallbackImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fallback image: %w", err)
	}

	txManager := infra.NewTxManager(db)

	tagRepo := infra.NewTagRepository()
	userRepo := infra.NewUserRepository()
	themeRepo := infra.NewThemeRepository()
	iconRepo := infra.NewIconRepository()

	tagUsecase := usecase.NewTagUsecase(txManager, tagRepo)
	themeUsecase := usecase.NewThemeUsecase(txManager, userRepo, themeRepo)
	userUsecase := usecase.NewUserUsecase(txManager, userRepo, themeRepo, iconRepo, fallbackIcon)
	iconUsecase := usecase.NewIconUsecase(txManager, userRepo, iconRepo)

	return &handlers{
		tag:   handler.NewTagHandler(tagUsecase),
		theme: handler.NewThemeHandler(themeUsecase),
		user:  handler.NewUserHandler(userUsecase),
		icon:  handler.NewIconHandler(iconUsecase, fallbackImagePath),
	}, nil
}
