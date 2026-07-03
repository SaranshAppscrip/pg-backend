package logger

import (
	"log/slog"

	"github.com/nivas/server/pkg/apperror"
)

// AppErrorAttrs returns slog attributes for an application error.
func AppErrorAttrs(err *apperror.AppError) []any {
	attrs := []any{
		"error_code", err.Code,
		"error_message", err.Message,
		"http_status", err.HTTPStatus,
	}
	if len(err.Details) > 0 {
		attrs = append(attrs, "error_details", err.Details)
	}
	if err.Err != nil {
		attrs = append(attrs, "cause", err.Err.Error())
	}
	return attrs
}

// LogAppError logs an AppError at warn (4xx) or error (5xx) level.
func LogAppError(log *slog.Logger, msg string, err *apperror.AppError) {
	attrs := AppErrorAttrs(err)
	if err.HTTPStatus >= 500 {
		log.Error(msg, attrs...)
		return
	}
	log.Warn(msg, attrs...)
}
