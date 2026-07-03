package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nivas/server/internal/config"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/pkg/apperror"
)

type TokenService struct {
	secret    []byte
	staffTTL  time.Duration
	tenantTTL time.Duration
}

func NewTokenService(cfg config.JWTConfig) *TokenService {
	return &TokenService{
		secret:    []byte(cfg.Secret),
		staffTTL:  cfg.StaffTTL,
		tenantTTL: cfg.TenantTTL,
	}
}

type claims struct {
	domain.AuthClaims
	jwt.RegisteredClaims
}

func (s *TokenService) GenerateStaffToken(staff *domain.Staff) (string, error) {
	return s.sign(domain.AuthClaims{
		Type:           domain.TokenTypeStaff,
		OrganizationID: staff.OrganizationID,
		UserID:         staff.ID,
		Email:          staff.Email,
		IsOwner:        staff.IsOwner,
	}, s.staffTTL)
}

func (s *TokenService) GenerateTenantToken(tenant *domain.Tenant) (string, error) {
	return s.sign(domain.AuthClaims{
		Type:           domain.TokenTypeTenant,
		OrganizationID: tenant.OrganizationID,
		UserID:         tenant.ID,
		Email:          tenant.Email,
	}, s.tenantTTL)
}

func (s *TokenService) Parse(tokenStr string) (*domain.AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, apperror.Unauthorized("invalid or expired token")
	}

	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		return nil, apperror.Unauthorized("invalid token")
	}

	return &c.AuthClaims, nil
}

func (s *TokenService) sign(authClaims domain.AuthClaims, ttl time.Duration) (string, error) {
	now := time.Now()
	c := claims{
		AuthClaims: authClaims,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "nivas",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", apperror.Internal("sign token", err)
	}
	return signed, nil
}
