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
	ContextKeyClaims   = "claims"
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
		c.Next()

		log.Info("request",
			"request_id", c.GetString(ContextKeyRequestID),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

func Recovery(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					"request_id", c.GetString(ContextKeyRequestID),
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
		token := extractBearer(c.GetHeader("Authorization"))
		if token == "" {
			response.Error(c, apperror.Unauthorized("missing authorization token"))
			c.Abort()
			return
		}

		claims, err := tokens.Parse(token)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}

		if claims.Type != expected {
			response.Error(c, apperror.Forbidden("invalid token type"))
			c.Abort()
			return
		}

		c.Set(ContextKeyClaims, claims)
		ctx := logger.WithContext(c.Request.Context(), slog.Default())
		c.Request = c.Request.WithContext(ctx)
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
