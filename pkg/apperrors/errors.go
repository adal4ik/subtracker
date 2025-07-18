package apperrors

import (
	"fmt"
	"net/http"
)

// AppError - кастомная структура для ошибок.
type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("AppError: %s (original: %v)", e.Message, e.Err)
	}
	return fmt.Sprintf("AppError: %s", e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func NewNotFound(message string, err error) *AppError {
	return New(http.StatusNotFound, message, err)
}

func NewBadRequest(message string, err error) *AppError {
	return New(http.StatusBadRequest, message, err)
}

func NewInternalServerError(message string, err error) *AppError {
	return New(http.StatusInternalServerError, message, err)
}
