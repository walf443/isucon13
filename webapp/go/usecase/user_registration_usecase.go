package usecase

import (
	"context"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

// UserRegistrationUsecase はユーザの登録を扱う。
type UserRegistrationUsecase interface {
	// Register はユーザを登録し、サブドメインの DNS レコードを登録して、テーマ・アイコンを含めたユーザを返す。
	// 予約済みのユーザ名の場合 *ReservedUsernameError を返す。
	Register(ctx context.Context, input RegisterUserInput) (*domain.User, error)
}

// RegisterUserInput は登録するユーザの内容。
type RegisterUserInput struct {
	Name        string
	DisplayName string
	Description string
	// Password はハッシュ化する前のパスワード。
	Password string
	DarkMode bool
}

type userRegistrationUsecase struct {
	txManager    repository.TxManager
	userRepo     repository.UserRepository
	themeRepo    repository.ThemeRepository
	dnsRegistrar DNSRecordRegistrar
}

func NewUserRegistrationUsecase(txManager repository.TxManager, userRepo repository.UserRepository, themeRepo repository.ThemeRepository, dnsRegistrar DNSRecordRegistrar) UserRegistrationUsecase {
	return &userRegistrationUsecase{txManager: txManager, userRepo: userRepo, themeRepo: themeRepo, dnsRegistrar: dnsRegistrar}
}

func (u *userRegistrationUsecase) Register(ctx context.Context, input RegisterUserInput) (*domain.User, error) {
	if reserved, ok := domain.FindReservedUsername(input.Name); ok {
		return nil, &ReservedUsernameError{Name: reserved}
	}

	hashedPassword, err := domain.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hashed password: %w", err)
	}

	var user *domain.User
	err = u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		userID, err := u.userRepo.Create(ctx, q, &domain.UserModel{
			Name:           input.Name,
			DisplayName:    input.DisplayName,
			Description:    input.Description,
			HashedPassword: hashedPassword,
		})
		if err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}

		if err := u.themeRepo.Create(ctx, q, &domain.ThemeModel{
			UserID:   userID,
			DarkMode: input.DarkMode,
		}); err != nil {
			return fmt.Errorf("failed to insert user theme: %w", err)
		}

		// 移行前と同じくトランザクション内で登録する (失敗した場合はユーザの登録もロールバックする)
		if err := u.dnsRegistrar.AddRecord(input.Name); err != nil {
			// 移行前はコマンドの出力とエラーをそのままレスポンスにしていたので、メッセージを付け足さない
			return err
		}

		user, err = u.userRepo.FindWithDetailsByID(ctx, q, userID)
		if err != nil {
			return fmt.Errorf("failed to fill user: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}
