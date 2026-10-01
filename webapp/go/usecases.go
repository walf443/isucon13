package main

import (
	"fmt"
	"os"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/infra/mysql"
	"github.com/isucon/isucon13/webapp/go/infra/powerdns"
	"github.com/isucon/isucon13/webapp/go/infra/script"
	"github.com/isucon/isucon13/webapp/go/interfaces/http/handler"
	"github.com/isucon/isucon13/webapp/go/usecase"
	"gorm.io/gorm"
)

// newUsecases は repository・usecase を組み立てる。
func newUsecases(db *gorm.DB, fallbackImagePath string, powerDNSSubdomainAddress string, logger usecase.Logger) (handler.Usecases, error) {
	fallbackIcon, err := os.ReadFile(fallbackImagePath)
	if err != nil {
		return handler.Usecases{}, fmt.Errorf("failed to read fallback image: %w", err)
	}
	// アイコン未登録のユーザのアイコンのハッシュは常に同じなので、起動時に 1 回だけ計算する
	defaultIconHash := domain.HashIcon(fallbackIcon)

	txManager := mysql.NewTxManager(db)

	tagRepo := mysql.NewTagRepository()
	userRepo := mysql.NewUserRepository()
	themeRepo := mysql.NewThemeRepository()
	iconRepo := mysql.NewIconRepository()
	livestreamRepo := mysql.NewLivestreamRepository()
	livestreamTagRepo := mysql.NewLivestreamTagRepository()
	reactionRepo := mysql.NewReactionRepository()
	livecommentRepo := mysql.NewLivecommentRepository()
	livecommentReportRepo := mysql.NewLivecommentReportRepository()
	ngWordRepo := mysql.NewNGWordRepository()
	viewerRepo := mysql.NewLivestreamViewersHistoryRepository()
	reservationSlotRepo := mysql.NewReservationSlotRepository()

	userFiller := usecase.NewUserFiller(themeRepo, iconRepo, defaultIconHash)
	livestreamFiller := usecase.NewLivestreamFiller(userRepo, livestreamTagRepo, tagRepo, userFiller)
	reactionFiller := usecase.NewReactionFiller(userRepo, livestreamRepo, userFiller, livestreamFiller)
	livecommentFiller := usecase.NewLivecommentFiller(userRepo, livestreamRepo, userFiller, livestreamFiller)
	livecommentReportFiller := usecase.NewLivecommentReportFiller(userRepo, livecommentRepo, userFiller, livecommentFiller)

	tagUsecase := usecase.NewTagUsecase(txManager, tagRepo)
	themeUsecase := usecase.NewThemeUsecase(txManager, userRepo, themeRepo)
	dnsRegistrar := powerdns.NewDNSRecordRegistrar(powerDNSSubdomainAddress)

	userUsecase := usecase.NewUserUsecase(txManager, userRepo, userFiller)
	userRegistrationUsecase := usecase.NewUserRegistrationUsecase(txManager, userRepo, themeRepo, dnsRegistrar, userFiller)
	iconUsecase := usecase.NewIconUsecase(txManager, userRepo, iconRepo)
	livestreamUsecase := usecase.NewLivestreamUsecase(txManager, userRepo, tagRepo, livestreamRepo, livestreamTagRepo, livestreamFiller)
	livestreamReservationUsecase := usecase.NewLivestreamReservationUsecase(txManager, livestreamRepo, livestreamTagRepo, reservationSlotRepo, livestreamFiller, logger)
	reactionUsecase := usecase.NewReactionUsecase(txManager, reactionRepo, reactionFiller)
	livecommentUsecase := usecase.NewLivecommentUsecase(txManager, livestreamRepo, livecommentRepo, ngWordRepo, livecommentFiller, logger)
	livecommentReportUsecase := usecase.NewLivecommentReportUsecase(txManager, livestreamRepo, livecommentRepo, livecommentReportRepo, livecommentReportFiller)
	ngWordUsecase := usecase.NewNGWordUsecase(txManager, livestreamRepo, livecommentRepo, ngWordRepo)
	viewerUsecase := usecase.NewLivestreamViewerUsecase(txManager, viewerRepo)
	statisticsUsecase := usecase.NewStatisticsUsecase(txManager, userRepo, livestreamRepo, livecommentRepo, reactionRepo, viewerRepo, livecommentReportRepo)
	paymentUsecase := usecase.NewPaymentUsecase(txManager, livecommentRepo)
	initializeUsecase := usecase.NewInitializeUsecase(script.NewInitializer("../sql/init.sh"), logger)

	return handler.Usecases{
		Tag:                   tagUsecase,
		Theme:                 themeUsecase,
		User:                  userUsecase,
		UserRegistration:      userRegistrationUsecase,
		Icon:                  iconUsecase,
		Livestream:            livestreamUsecase,
		LivestreamReservation: livestreamReservationUsecase,
		Reaction:              reactionUsecase,
		Livecomment:           livecommentUsecase,
		LivecommentReport:     livecommentReportUsecase,
		NGWord:                ngWordUsecase,
		LivestreamViewer:      viewerUsecase,
		Statistics:            statisticsUsecase,
		Payment:               paymentUsecase,
		Initialize:            initializeUsecase,
	}, nil
}
