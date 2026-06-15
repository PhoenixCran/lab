package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"

	"lab/internal/model"
)

type BookRepository interface {
	Create(ctx context.Context, book *model.Book) error
	GetByID(ctx context.Context, id int64) (*model.Book, error)
	List(ctx context.Context) ([]model.Book, error)
	Update(ctx context.Context, book *model.Book) error
	Delete(ctx context.Context, id int64) (bool, error)
}

type bookRepoImpl struct {
	db *sql.DB
}

func NewBookRepo(db *sql.DB) BookRepository {
	return &bookRepoImpl{db: db}
}

func (r *bookRepoImpl) Create(ctx context.Context, book *model.Book) error {
	query := `INSERT INTO books (title, author, isbn, category, status) VALUES (?, ?, ?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, book.Title, book.Author, book.ISBN, book.Category, book.Status)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return fmt.Errorf("isbn already exists: %w", err)
		}
		return fmt.Errorf("create book: %w", err)
	}
	id, _ := result.LastInsertId()

	// Read back to get DB-generated defaults (created_at, updated_at)
	created, err := r.GetByID(ctx, id)
	if err != nil || created == nil {
		return fmt.Errorf("create book: readback failed: %w", err)
	}
	*book = *created
	return nil
}

func (r *bookRepoImpl) GetByID(ctx context.Context, id int64) (*model.Book, error) {
	query := `SELECT id, title, author, isbn, category, status, created_at, updated_at FROM books WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)
	book := &model.Book{}
	err := row.Scan(&book.ID, &book.Title, &book.Author, &book.ISBN, &book.Category, &book.Status, &book.CreatedAt, &book.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get book by id: %w", err)
	}
	return book, nil
}

func (r *bookRepoImpl) List(ctx context.Context) ([]model.Book, error) {
	query := `SELECT id, title, author, isbn, category, status, created_at, updated_at FROM books ORDER BY id DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list books: %w", err)
	}
	defer rows.Close()

	books := make([]model.Book, 0)
	for rows.Next() {
		var b model.Book
		if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Category, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan book: %w", err)
		}
		books = append(books, b)
	}
	return books, rows.Err()
}

func (r *bookRepoImpl) Update(ctx context.Context, book *model.Book) error {
	query := `UPDATE books SET title=?, author=?, isbn=?, category=?, status=? WHERE id=?`
	result, err := r.db.ExecContext(ctx, query, book.Title, book.Author, book.ISBN, book.Category, book.Status, book.ID)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return fmt.Errorf("isbn already exists: %w", err)
		}
		return fmt.Errorf("update book: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil
	}
	return nil
}

func (r *bookRepoImpl) Delete(ctx context.Context, id int64) (bool, error) {
	query := `DELETE FROM books WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("delete book: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}