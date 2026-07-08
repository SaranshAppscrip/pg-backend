package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nivas/server/pkg/apperror"
)

// OptionalPropertyID reads property_id from query string. Nil means all properties.
func OptionalPropertyID(c *gin.Context) (*uuid.UUID, error) {
	raw := c.Query("property_id")
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, apperror.BadRequest("invalid property_id")
	}
	return &id, nil
}

// RequiredPropertyID reads property_id from query string (required for room create).
func RequiredPropertyID(c *gin.Context) (uuid.UUID, error) {
	id, err := OptionalPropertyID(c)
	if err != nil {
		return uuid.Nil, err
	}
	if id == nil {
		return uuid.Nil, apperror.BadRequest("property_id is required")
	}
	return *id, nil
}
