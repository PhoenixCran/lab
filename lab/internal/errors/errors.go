package errors

import "fmt"

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string { return e.Message }

func (e *AppError) Unwrap() error { return e.Err }

func ErrNotFound(entity string, id int64) *AppError {
	return &AppError{
		Code:    404,
		Message: fmt.Sprintf("%s with id %d not found", entity, id),
	}
}

func ErrValidation(field, msg string) *AppError {
	return &AppError{
		Code:    400,
		Message: fmt.Sprintf("validation error: %s — %s", field, msg),
	}
}

func ErrConflict(field, value string) *AppError {
	return &AppError{
		Code:    409,
		Message: fmt.Sprintf("%s '%s' already exists", field, value),
	}
}

func ErrInternal(err error) *AppError {
	return &AppError{
		Code:    500,
		Message: "internal server error",
		Err:     err,
	}
}