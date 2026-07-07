package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/pkg/apperror"
)

func (s *Store) CreateRefreshToken(
	ctx context.Context,
	userType domain.TokenType,
	userID, orgID uuid.UUID,
	tokenHash string,
	expiresAt time.Time,
) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (user_type, user_id, organization_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, string(userType), userID, orgID, tokenHash, expiresAt)
	return mapPgError(err, "")
}

func (s *Store) GetValidRefreshTokenByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_type, user_id, organization_id
		FROM refresh_tokens
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > NOW()
	`, tokenHash).Scan(&rt.ID, &rt.UserType, &rt.UserID, &rt.OrganizationID)
	if err != nil {
		return nil, mapPgError(err, "invalid or expired refresh token")
	}
	return &rt, nil
}

func (s *Store) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	if err != nil {
		return mapPgError(err, "")
	}
	if tag.RowsAffected() == 0 {
		return apperror.Unauthorized("invalid or expired refresh token")
	}
	return nil
}

func (s *Store) RevokeAllRefreshTokensForUser(ctx context.Context, userType domain.TokenType, userID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_type = $1 AND user_id = $2 AND revoked_at IS NULL
	`, string(userType), userID)
	return mapPgError(err, "")
}
