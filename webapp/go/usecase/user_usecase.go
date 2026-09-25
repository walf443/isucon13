package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/isucon/isucon13/webapp/go/domain/model"
	"github.com/isucon/isucon13/webapp/go/domain/repository"
)

type UserUsecase interface {
	// FindByID はユーザが存在しない場合 ErrUserNotFound を返す。
	FindByID(ctx context.Context, id model.UserID) (*model.User, error)
	// FindByName はユーザが存在しない場合 ErrUserNotFound を返す。
	FindByName(ctx context.Context, name string) (*model.User, error)
	// Register はユーザを登録し、サブドメインの DNS レコードを登録して、テーマ・アイコンを含めたユーザを返す。
	// 予約済みのユーザ名の場合 ErrReservedUsername を返す。
	Register(ctx context.Context, input RegisterUserInput) (*model.User, error)
	// Login はユーザ名とパスワードを検証し、ユーザを返す。
	// ユーザが存在しないかパスワードが違う場合 ErrInvalidCredentials を返す。
	Login(ctx context.Context, username string, password string) (*model.UserModel, error)
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

type userUsecase struct {
	txManager    repository.TxManager
	userRepo     repository.UserRepository
	themeRepo    repository.ThemeRepository
	dnsRegistrar DNSRecordRegistrar
}

func NewUserUsecase(txManager repository.TxManager, userRepo repository.UserRepository, themeRepo repository.ThemeRepository, dnsRegistrar DNSRecordRegistrar) UserUsecase {
	return &userUsecase{txManager: txManager, userRepo: userRepo, themeRepo: themeRepo, dnsRegistrar: dnsRegistrar}
}

func (u *userUsecase) FindByID(ctx context.Context, id model.UserID) (*model.User, error) {
	return u.findUser(ctx, func(q repository.Querier) (*model.User, error) {
		return u.userRepo.FindWithDetailsByID(ctx, q, id)
	})
}

func (u *userUsecase) FindByName(ctx context.Context, name string) (*model.User, error) {
	return u.findUser(ctx, func(q repository.Querier) (*model.User, error) {
		return u.userRepo.FindWithDetailsByName(ctx, q, name)
	})
}

func (u *userUsecase) findUser(ctx context.Context, find func(q repository.Querier) (*model.User, error)) (*model.User, error) {
	var user *model.User
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		user, err = find(q)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userUsecase) Register(ctx context.Context, input RegisterUserInput) (*model.User, error) {
	if input.Name == "pipe" {
		return nil, ErrReservedUsername
	}

	hashedPassword, err := model.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hashed password: %w", err)
	}

	var user *model.User
	err = u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		userID, err := u.userRepo.Create(ctx, q, &model.UserModel{
			Name:           input.Name,
			DisplayName:    input.DisplayName,
			Description:    input.Description,
			HashedPassword: hashedPassword,
		})
		if err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}

		if err := u.themeRepo.Create(ctx, q, &model.ThemeModel{
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

func (u *userUsecase) Login(ctx context.Context, username string, password string) (*model.UserModel, error) {
	var user *model.UserModel
	err := u.txManager.RunInTx(ctx, func(q repository.Querier) error {
		var err error
		// usernameはUNIQUEなので、whereで一意に特定できる
		user, err = u.userRepo.FindByName(ctx, q, username)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrInvalidCredentials
		}
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 移行前と同じく、トランザクションをコミットしてからパスワードを検証する
	matched, err := user.HashedPassword.Matches(password)
	if err != nil {
		return nil, fmt.Errorf("failed to compare hash and password: %w", err)
	}
	if !matched {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}
