// 用户服务层
// 实现用户相关的业务逻辑
// 包含用户验证、创建、查询、更新和删除等核心业务规则
package service

import (
	"context"
	"net/mail"
	"strings"

	appErrors "lab/internal/errors"
	"lab/internal/model"
	"lab/internal/repository"
)

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name"`
	Email *string `json:"email"`
	Phone *string `json:"phone"`
}

type UserService interface {
	Create(ctx context.Context, req CreateUserRequest) (*model.User, error)
	GetByID(ctx context.Context, id int64) (*model.User, error)
	List(ctx context.Context) ([]model.User, error)
	Update(ctx context.Context, id int64, req UpdateUserRequest) (*model.User, error)
	Delete(ctx context.Context, id int64) error
}

type userServiceImpl struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userServiceImpl{repo: repo}
}

func (s *userServiceImpl) Create(ctx context.Context, req CreateUserRequest) (*model.User, error) {
	if err := validateUserNameEmail(req.Name, req.Email); err != nil {
		return nil, err
	}

	user := &model.User{
		Name:  strings.TrimSpace(req.Name),
		Email: strings.TrimSpace(req.Email),
		Phone: strings.TrimSpace(req.Phone),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, appErrors.ErrConflict("email", req.Email)
		}
		return nil, appErrors.ErrInternal(err)
	}
	return user, nil
}

func (s *userServiceImpl) GetByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, appErrors.ErrInternal(err)
	}
	if user == nil {
		return nil, appErrors.ErrNotFound("user", id)
	}
	return user, nil
}

func (s *userServiceImpl) List(ctx context.Context) ([]model.User, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, appErrors.ErrInternal(err)
	}
	return users, nil
}

func (s *userServiceImpl) Update(ctx context.Context, id int64, req UpdateUserRequest) (*model.User, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, appErrors.ErrInternal(err)
	}
	if existing == nil {
		return nil, appErrors.ErrNotFound("user", id)
	}

	if req.Name != nil {
		existing.Name = strings.TrimSpace(*req.Name)
	}
	if req.Email != nil {
		existing.Email = strings.TrimSpace(*req.Email)
	}
	if req.Phone != nil {
		existing.Phone = strings.TrimSpace(*req.Phone)
	}

	if err := validateUserNameEmail(existing.Name, existing.Email); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, appErrors.ErrConflict("email", existing.Email)
		}
		return nil, appErrors.ErrInternal(err)
	}

	// Read back to get DB-updated fields (updated_at)
	return s.GetByID(ctx, id)
}

func (s *userServiceImpl) Delete(ctx context.Context, id int64) error {
	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return appErrors.ErrInternal(err)
	}
	if !deleted {
		return appErrors.ErrNotFound("user", id)
	}
	return nil
}

func validateUserNameEmail(name, email string) error {
	if strings.TrimSpace(name) == "" {
		return appErrors.ErrValidation("name", "name is required")
	}
	if strings.TrimSpace(email) == "" {
		return appErrors.ErrValidation("email", "email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return appErrors.ErrValidation("email", "invalid email format")
	}
	return nil
}