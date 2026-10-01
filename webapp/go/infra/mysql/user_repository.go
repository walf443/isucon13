package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/isucon/isucon13/webapp/go/domain"
	"github.com/isucon/isucon13/webapp/go/usecase/repository"
)

type userRepository struct{}

func NewUserRepository() repository.UserRepository {
	return &userRepository{}
}

func (r *userRepository) FindIDByName(ctx context.Context, q repository.Querier, name string) (domain.UserID, error) {
	var id domain.UserID
	err := q.GetContext(ctx, &id, "SELECT id FROM users WHERE name = ?", name)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, repository.ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *userRepository) FindByName(ctx context.Context, q repository.Querier, name string) (*domain.User, error) {
	var user domain.User
	err := q.GetContext(ctx, &user, "SELECT id, name, display_name, description, password FROM users WHERE name = ?", name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, q repository.Querier, id domain.UserID) (*domain.User, error) {
	var user domain.User
	err := q.GetContext(ctx, &user, "SELECT id, name, display_name, description, password FROM users WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context, q repository.Querier) ([]*domain.User, error) {
	var users []*domain.User
	if err := q.SelectContext(ctx, &users, "SELECT id, name, display_name, description, password FROM users"); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Create(ctx context.Context, q repository.Querier, user *domain.User) (domain.UserID, error) {
	result, err := q.ExecContext(ctx, "INSERT INTO users (name, display_name, description, password) VALUES(?, ?, ?, ?)", user.Name, user.DisplayName, user.Description, user.HashedPassword)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return domain.UserID(id), nil
}

func (r *userRepository) FindAllByIDs(ctx context.Context, q repository.Querier, ids []domain.UserID) ([]*domain.User, error) {
	return selectIn[domain.User](ctx, q, "SELECT id, name, display_name, description, password FROM users WHERE id IN (?)", ids)
}
