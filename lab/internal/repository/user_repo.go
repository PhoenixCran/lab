package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"

	"lab/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	List(ctx context.Context) ([]model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id int64) (bool, error)
}

type userRepoImpl struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) UserRepository {
	return &userRepoImpl{db: db}
}

func (r *userRepoImpl) Create(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (name, email, phone) VALUES (?, ?, ?)`
	result, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.Phone)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return fmt.Errorf("email already exists: %w", err)
		}
		return fmt.Errorf("create user: %w", err)
	}
	id, _ := result.LastInsertId()

	// Read back to get DB-generated defaults (created_at, updated_at)
	created, err := r.GetByID(ctx, id)
	if err != nil || created == nil {
		return fmt.Errorf("create user: readback failed: %w", err)
	}
	*user = *created
	return nil
}

func (r *userRepoImpl) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `SELECT id, name, email, phone, created_at, updated_at FROM users WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)
	user := &model.User{}
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Phone, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (r *userRepoImpl) List(ctx context.Context) ([]model.User, error) {
	query := `SELECT id, name, email, phone, created_at, updated_at FROM users ORDER BY id DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Phone, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *userRepoImpl) Update(ctx context.Context, user *model.User) error {
	query := `UPDATE users SET name=?, email=?, phone=? WHERE id=?`
	result, err := r.db.ExecContext(ctx, query, user.Name, user.Email, user.Phone, user.ID)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			return fmt.Errorf("email already exists: %w", err)
		}
		return fmt.Errorf("update user: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil
	}
	return nil
}

func (r *userRepoImpl) Delete(ctx context.Context, id int64) (bool, error) {
	query := `DELETE FROM users WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return false, fmt.Errorf("delete user: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}