package main

import (
	"fmt"
	"os"

	infra "github.com/isucon/isucon13/webapp/go/infra/mysql"
	"github.com/isucon/isucon13/webapp/go/infra/powerdns"
	"github.com/isucon/isucon13/webapp/go/infra/script"
	"github.com/isucon/isucon13/webapp/go/interfaces/http/handler"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"github.com/jmoiron/sqlx"
)

// newUsecases は repository・usecase を組み立てる。
func newUsecases(db *sqlx.DB, fallbackImagePath string, powerDNSSubdomainAddress string, logger usecase.Logger) (handler.Usecases, error) {
	fallbackIcon, err := os.ReadFile(fallbackImagePath)
	if err != nil {
		return handler.Usecases{}, fmt.Errorf("failed to read fallback image: %w", err)
	}

	txManager := infra.NewTxManager(db)

	tagRepo := infra.NewTagRepository()
	userRepo := infra.NewUserRepository(fallbackIcon)
	themeRepo := infra.NewThemeRepository()
	iconRepo := infra.NewIconRepository()
	livestreamRepo := infra.NewLivestreamRepository(fallbackIcon)
	reactionRepo := infra.NewReactionRepository(fallbackIcon)
	livecommentRepo := infra.NewLivecommentRepository(fallbackIcon)
	livecommentReportRepo := infra.NewLivecommentReportRepository(fallbackIcon)
	ngWordRepo := infra.NewNGWordRepository()
	viewerRepo := infra.NewLivestreamViewersHistoryRepository()
	reservationSlotRepo := infra.NewReservationSlotRepository()

	tagUsecase := usecase.NewTagUsecase(txManager, tagRepo)
	themeUsecase := usecase.NewThemeUsecase(txManager, userRepo, themeRepo)
	dnsRegistrar := powerdns.NewDNSRecordRegistrar(powerDNSSubdomainAddress)

	userUsecase := usecase.NewUserUsecase(txManager, userRepo, themeRepo, dnsRegistrar)
	iconUsecase := usecase.NewIconUsecase(txManager, userRepo, iconRepo)
	livestreamUsecase := usecase.NewLivestreamUsecase(txManager, userRepo, tagRepo, livestreamRepo, reservationSlotRepo, logger)
	reactionUsecase := usecase.NewReactionUsecase(txManager, reactionRepo)
	livecommentUsecase := usecase.NewLivecommentUsecase(txManager, livestreamRepo, livecommentRepo, livecommentReportRepo, ngWordRepo, logger)
	ngWordUsecase := usecase.NewNGWordUsecase(txManager, livestreamRepo, livecommentRepo, ngWordRepo)
	viewerUsecase := usecase.NewLivestreamViewerUsecase(txManager, viewerRepo)
	statisticsUsecase := usecase.NewStatisticsUsecase(txManager, userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo, livecommentReportRepo)
	paymentUsecase := usecase.NewPaymentUsecase(txManager, livecommentRepo)
	initializeUsecase := usecase.NewInitializeUsecase(script.NewInitializer("../sql/init.sh"), logger)

	return handler.Usecases{
		Tag:              tagUsecase,
		Theme:            themeUsecase,
		User:             userUsecase,
		Icon:             iconUsecase,
		Livestream:       livestreamUsecase,
		Reaction:         reactionUsecase,
		Livecomment:      livecommentUsecase,
		NGWord:           ngWordUsecase,
		LivestreamViewer: viewerUsecase,
		Statistics:       statisticsUsecase,
		Payment:          paymentUsecase,
		Initialize:       initializeUsecase,
	}, nil
}
