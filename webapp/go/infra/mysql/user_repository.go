package mysql

import (
	"context"
	"errors"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
	"gorm.io/gorm"
)

// userRow は users テーブルの行。
type userRow struct {
	ID          int64  `gorm:"column:id;primaryKey"`
	Name        string `gorm:"column:name"`
	DisplayName string `gorm:"column:display_name"`
	Description string `gorm:"column:description"`
	Password    string `gorm:"column:password"`
}

func (userRow) TableName() string { return "users" }

func (r *userRow) toDomain() *domain.User {
	return &domain.User{
		ID:             domain.UserID(r.ID),
		Name:           r.Name,
		DisplayName:    r.DisplayName,
		Description:    r.Description,
		HashedPassword: domain.HashedPassword(r.Password),
	}
}

type userRepository struct{}

func NewUserRepository() repository.UserRepository {
	return &userRepository{}
}

func (r *userRepository) FindIDByName(ctx context.Context, q repository.Querier, name string) (domain.UserID, error) {
	var row userRow
	err := dbOf(ctx, q).Select("id").Where("name = ?", name).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, repository.ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	return domain.UserID(row.ID), nil
}

func (r *userRepository) FindByName(ctx context.Context, q repository.Querier, name string) (*domain.User, error) {
	var row userRow
	// ユーザ名は UNIQUE なので、where で一意に特定できる
	err := dbOf(ctx, q).Select("id, name, display_name, description, password").Where("name = ?", name).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return row.toDomain(), nil
}

func (r *userRepository) FindByID(ctx context.Context, q repository.Querier, id domain.UserID) (*domain.User, error) {
	var row userRow
	err := dbOf(ctx, q).Select("id, name, display_name, description, password").Where("id = ?", id).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return row.toDomain(), nil
}

func (r *userRepository) FindAll(ctx context.Context, q repository.Querier) ([]*domain.User, error) {
	var rows []*userRow
	if err := dbOf(ctx, q).Select("id, name, display_name, description, password").Find(&rows).Error; err != nil {
		return nil, err
	}
	return mapRows(rows, (*userRow).toDomain), nil
}

func (r *userRepository) Create(ctx context.Context, q repository.Querier, user *domain.User) (domain.UserID, error) {
	row := userRow{
		Name:        user.Name,
		DisplayName: user.DisplayName,
		Description: user.Description,
		Password:    string(user.HashedPassword),
	}
	// ユーザ名が重複した場合は、ドライバのエラー (Error 1062) がそのまま返る (GORM のエラーの翻訳はしない)
	if err := dbOf(ctx, q).Create(&row).Error; err != nil {
		return 0, err
	}
	return domain.UserID(row.ID), nil
}

func (r *userRepository) FindAllByIDs(ctx context.Context, q repository.Querier, ids []domain.UserID) ([]*domain.User, error) {
	rows, err := findIn(ids, func(chunk []domain.UserID) ([]*userRow, error) {
		var rows []*userRow
		err := dbOf(ctx, q).Select("id, name, display_name, description, password").Where("id IN ?", chunk).Find(&rows).Error
		return rows, err
	})
	if err != nil {
		return nil, err
	}
	return mapRows(rows, (*userRow).toDomain), nil
}
