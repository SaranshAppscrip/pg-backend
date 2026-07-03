package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// Code is a stable machine-readable error identifier.
type Code string

const (
	CodeInternal       Code = "INTERNAL_ERROR"
	CodeNotFound       Code = "NOT_FOUND"
	CodeBadRequest     Code = "BAD_REQUEST"
	CodeUnauthorized   Code = "UNAUTHORIZED"
	CodeForbidden      Code = "FORBIDDEN"
	CodeConflict       Code = "CONFLICT"
	CodeValidation     Code = "VALIDATION_ERROR"
)

// AppError is the application's typed error with HTTP mapping.
type AppError struct {
	Code       Code              `json:"code"`
	Message    string            `json:"message"`
	HTTPStatus int               `json:"-"`
	Details    []map[string]string `json:"details,omitempty"`
	Err        error             `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates an AppError.
func New(code Code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Wrap wraps an underlying error with an AppError.
func Wrap(err error, code Code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// WithDetails adds field-level validation details.
func (e *AppError) WithDetails(details []map[string]string) *AppError {
	e.Details = details
	return e
}

// Common constructors
func Internal(msg string, err error) *AppError {
	return Wrap(err, CodeInternal, msg, http.StatusInternalServerError)
}

func NotFound(msg string) *AppError {
	return New(CodeNotFound, msg, http.StatusNotFound)
}

func BadRequest(msg string) *AppError {
	return New(CodeBadRequest, msg, http.StatusBadRequest)
}

func Unauthorized(msg string) *AppError {
	return New(CodeUnauthorized, msg, http.StatusUnauthorized)
}

func Forbidden(msg string) *AppError {
	return New(CodeForbidden, msg, http.StatusForbidden)
}

func Conflict(msg string) *AppError {
	return New(CodeConflict, msg, http.StatusConflict)
}

func Validation(msg string, details []map[string]string) *AppError {
	return New(CodeValidation, msg, http.StatusBadRequest).WithDetails(details)
}

// Is checks if err is an AppError with the given code.
func Is(err error, code Code) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// As extracts an AppError from err.
func As(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// HTTPStatus returns the HTTP status for err, defaulting to 500.
func HTTPStatus(err error) int {
	if appErr, ok := As(err); ok {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}
