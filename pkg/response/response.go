package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nivas/server/pkg/apperror"
	"github.com/nivas/server/pkg/logger"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    apperror.Code       `json:"code"`
	Message string              `json:"message"`
	Details []map[string]string `json:"details,omitempty"`
}

// JSON sends a successful JSON response.
func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}

// OK sends 200 with data.
func OK(c *gin.Context, data any) {
	JSON(c, http.StatusOK, data)
}

// Created sends 201 with data.
func Created(c *gin.Context, data any) {
	JSON(c, http.StatusCreated, data)
}

// NoContent sends 204.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error maps application errors to a consistent JSON error response.
func Error(c *gin.Context, err error) {
	_ = c.Error(err)

	log := logger.FromContext(c.Request.Context())
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		logger.LogAppError(log, "handler error", appErr)
		c.JSON(appErr.HTTPStatus, errorBody{
			Error: errorDetail{
				Code:    appErr.Code,
				Message: appErr.Message,
				Details: appErr.Details,
			},
		})
		return
	}

	log.Error("unhandled error", "error", err.Error())
	c.JSON(http.StatusInternalServerError, errorBody{
		Error: errorDetail{
			Code:    apperror.CodeInternal,
			Message: "An unexpected error occurred",
		},
	})
}
