// 图书服务层
// 实现图书相关的业务逻辑
// 包含图书验证、创建、查询、更新和删除等核心业务规则
package service

import (
	"context"
	"strings"

	appErrors "lab/internal/errors"
	"lab/internal/model"
	"lab/internal/repository"
)

type CreateBookRequest struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	ISBN     string `json:"isbn"`
	Category string `json:"category"`
}

type UpdateBookRequest struct {
	Title    *string `json:"title"`
	Author   *string `json:"author"`
	ISBN     *string `json:"isbn"`
	Category *string `json:"category"`
	Status   *string `json:"status"`
}

type BookService interface {
	Create(ctx context.Context, req CreateBookRequest) (*model.Book, error)
	GetByID(ctx context.Context, id int64) (*model.Book, error)
	List(ctx context.Context) ([]model.Book, error)
	Update(ctx context.Context, id int64, req UpdateBookRequest) (*model.Book, error)
	Delete(ctx context.Context, id int64) error
}

type bookServiceImpl struct {
	repo repository.BookRepository
}

func NewBookService(repo repository.BookRepository) BookService {
	return &bookServiceImpl{repo: repo}
}

func (s *bookServiceImpl) Create(ctx context.Context, req CreateBookRequest) (*model.Book, error) {
	if err := validateBook(req.Title, req.Author, req.ISBN, req.Category); err != nil {
		return nil, err
	}

	book := &model.Book{
		Title:    strings.TrimSpace(req.Title),
		Author:   strings.TrimSpace(req.Author),
		ISBN:     strings.TrimSpace(req.ISBN),
		Category: strings.TrimSpace(req.Category),
		Status:   "available",
	}

	if err := s.repo.Create(ctx, book); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, appErrors.ErrConflict("isbn", req.ISBN)
		}
		return nil, appErrors.ErrInternal(err)
	}
	return book, nil
}

func (s *bookServiceImpl) GetByID(ctx context.Context, id int64) (*model.Book, error) {
	book, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, appErrors.ErrInternal(err)
	}
	if book == nil {
		return nil, appErrors.ErrNotFound("book", id)
	}
	return book, nil
}

func (s *bookServiceImpl) List(ctx context.Context) ([]model.Book, error) {
	books, err := s.repo.List(ctx)
	if err != nil {
		return nil, appErrors.ErrInternal(err)
	}
	return books, nil
}

func (s *bookServiceImpl) Update(ctx context.Context, id int64, req UpdateBookRequest) (*model.Book, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, appErrors.ErrInternal(err)
	}
	if existing == nil {
		return nil, appErrors.ErrNotFound("book", id)
	}

	if req.Title != nil {
		existing.Title = strings.TrimSpace(*req.Title)
	}
	if req.Author != nil {
		existing.Author = strings.TrimSpace(*req.Author)
	}
	if req.ISBN != nil {
		existing.ISBN = strings.TrimSpace(*req.ISBN)
	}
	if req.Category != nil {
		existing.Category = strings.TrimSpace(*req.Category)
	}
	if req.Status != nil {
		status := strings.TrimSpace(*req.Status)
		if status != "available" && status != "borrowed" {
			return nil, appErrors.ErrValidation("status", "must be 'available' or 'borrowed'")
		}
		existing.Status = status
	}

	if err := validateBook(existing.Title, existing.Author, existing.ISBN, existing.Category); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil, appErrors.ErrConflict("isbn", existing.ISBN)
		}
		return nil, appErrors.ErrInternal(err)
	}

	// Read back to get DB-updated fields (updated_at)
	return s.GetByID(ctx, id)
}

func (s *bookServiceImpl) Delete(ctx context.Context, id int64) error {
	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return appErrors.ErrInternal(err)
	}
	if !deleted {
		return appErrors.ErrNotFound("book", id)
	}
	return nil
}

func validateBook(title, author, isbn, category string) error {
	if strings.TrimSpace(title) == "" {
		return appErrors.ErrValidation("title", "title is required")
	}
	if strings.TrimSpace(author) == "" {
		return appErrors.ErrValidation("author", "author is required")
	}
	if strings.TrimSpace(isbn) == "" {
		return appErrors.ErrValidation("isbn", "isbn is required")
	}
	if strings.TrimSpace(category) == "" {
		return appErrors.ErrValidation("category", "category is required")
	}
	return nil
}