package middleware

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nivas/server/internal/auth"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/pkg/apperror"
	"github.com/nivas/server/pkg/logger"
	"github.com/nivas/server/pkg/response"
)

const (
	ContextKeyClaims    = "claims"
	ContextKeyRequestID = "request_id"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.New().String()
		}
		c.Set(ContextKeyRequestID, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.GetString(ContextKeyRequestID)

		reqLog := log.With(
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)
		ctx := logger.WithContext(c.Request.Context(), reqLog)
		c.Request = c.Request.WithContext(ctx)

		if c.Request.URL.RawQuery != "" {
			reqLog.Debug("incoming request", "query", c.Request.URL.RawQuery)
		}

		c.Next()

		status := c.Writer.Status()
		attrs := []any{
			"status", status,
			"duration_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
			"bytes_written", c.Writer.Size(),
		}

		if claims := GetClaims(c); claims != nil {
			attrs = append(attrs,
				"user_id", claims.UserID,
				"organization_id", claims.OrganizationID,
				"user_type", claims.Type,
			)
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, "handler_errors", c.Errors.String())
		}

		msg := "request completed"
		switch {
		case status >= 500:
			reqLog.Error(msg, attrs...)
		case status >= 400:
			reqLog.Warn(msg, attrs...)
		default:
			reqLog.Info(msg, attrs...)
		}
	}
}

func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					"request_id", c.GetString(ContextKeyRequestID),
					"path", c.Request.URL.Path,
					"panic", r,
				)
				response.Error(c, apperror.Internal("unexpected server error", nil))
				c.Abort()
			}
		}()
		c.Next()
	}
}

func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && isAllowedOrigin(origin, allowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func isAllowedOrigin(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == "*" || a == origin {
			return true
		}
	}
	return false
}

func StaffAuth(tokens *auth.TokenService) gin.HandlerFunc {
	return authMiddleware(tokens, domain.TokenTypeStaff)
}

func TenantAuth(tokens *auth.TokenService) gin.HandlerFunc {
	return authMiddleware(tokens, domain.TokenTypeTenant)
}

func authMiddleware(tokens *auth.TokenService, expected domain.TokenType) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.FromContext(c.Request.Context())

		token := extractBearer(c.GetHeader("Authorization"))
		if token == "" {
			log.Warn("auth rejected", "reason", "missing_token", "expected_type", expected)
			response.Error(c, apperror.Unauthorized("missing authorization token"))
			c.Abort()
			return
		}

		claims, err := tokens.Parse(token)
		if err != nil {
			log.Warn("auth rejected", "reason", "invalid_token", "expected_type", expected)
			response.Error(c, err)
			c.Abort()
			return
		}

		if claims.Type != expected {
			log.Warn("auth rejected",
				"reason", "invalid_token_type",
				"expected_type", expected,
				"actual_type", claims.Type,
				"user_id", claims.UserID,
			)
			response.Error(c, apperror.Forbidden("invalid token type"))
			c.Abort()
			return
		}

		c.Set(ContextKeyClaims, claims)

		reqLog := log.With(
			"user_id", claims.UserID,
			"organization_id", claims.OrganizationID,
			"user_type", claims.Type,
		)
		ctx := logger.WithContext(c.Request.Context(), reqLog)
		c.Request = c.Request.WithContext(ctx)

		log.Debug("auth ok", "user_id", claims.UserID, "user_type", claims.Type)
		c.Next()
	}
}

func extractBearer(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func GetClaims(c *gin.Context) *domain.AuthClaims {
	val, _ := c.Get(ContextKeyClaims)
	claims, _ := val.(*domain.AuthClaims)
	return claims
}

func GetOrganizationID(c *gin.Context) uuid.UUID {
	claims := GetClaims(c)
	if claims == nil {
		return uuid.Nil
	}
	return claims.OrganizationID
}

func GetUserID(c *gin.Context) uuid.UUID {
	claims := GetClaims(c)
	if claims == nil {
		return uuid.Nil
	}
	return claims.UserID
}

func GetUserType(c *gin.Context) domain.TokenType {
	claims := GetClaims(c)
	if claims == nil {
		return ""
	}
	return claims.Type
}

func IsStaffOwner(c *gin.Context) bool {
	claims := GetClaims(c)
	return claims != nil && claims.IsOwner
}
