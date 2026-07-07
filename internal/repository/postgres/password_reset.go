package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/pkg/apperror"
)

func (s *Store) UpdateStaffPassword(ctx context.Context, staffID uuid.UUID, passwordHash string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE staff SET password_hash = $1 WHERE id = $2
	`, passwordHash, staffID)
	if err != nil {
		return mapPgError(err, "")
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("staff not found")
	}
	return nil
}

func (s *Store) InvalidateUnusedPasswordResetTokens(ctx context.Context, staffID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE password_reset_tokens
		SET used_at = NOW()
		WHERE staff_id = $1 AND used_at IS NULL
	`, staffID)
	return mapPgError(err, "")
}

func (s *Store) CreatePasswordResetToken(ctx context.Context, staffID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO password_reset_tokens (staff_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, staffID, tokenHash, expiresAt)
	return mapPgError(err, "")
}

func (s *Store) GetValidPasswordResetByTokenHash(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	var staffID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT staff_id
		FROM password_reset_tokens
		WHERE token_hash = $1
		  AND used_at IS NULL
		  AND expires_at > NOW()
	`, tokenHash).Scan(&staffID)
	if err != nil {
		return uuid.Nil, mapPgError(err, "invalid or expired reset token")
	}
	return staffID, nil
}

func (s *Store) MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE password_reset_tokens
		SET used_at = NOW()
		WHERE token_hash = $1 AND used_at IS NULL
	`, tokenHash)
	if err != nil {
		return mapPgError(err, "")
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("invalid or expired reset token")
	}
	return nil
}

func (s *Store) UpdateTenantPassword(ctx context.Context, tenantID uuid.UUID, passwordHash string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE tenants SET password_hash = $1 WHERE id = $2
	`, passwordHash, tenantID)
	if err != nil {
		return mapPgError(err, "")
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("tenant not found")
	}
	return nil
}

func (s *Store) InvalidateUnusedTenantPasswordResetTokens(ctx context.Context, tenantID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE tenant_password_reset_tokens
		SET used_at = NOW()
		WHERE tenant_id = $1 AND used_at IS NULL
	`, tenantID)
	return mapPgError(err, "")
}

func (s *Store) CreateTenantPasswordResetToken(ctx context.Context, tenantID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO tenant_password_reset_tokens (tenant_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, tenantID, tokenHash, expiresAt)
	return mapPgError(err, "")
}

func (s *Store) GetValidTenantPasswordResetByTokenHash(ctx context.Context, tokenHash string) (uuid.UUID, error) {
	var tenantID uuid.UUID
	err := s.pool.QueryRow(ctx, `
		SELECT tenant_id
		FROM tenant_password_reset_tokens
		WHERE token_hash = $1
		  AND used_at IS NULL
		  AND expires_at > NOW()
	`, tokenHash).Scan(&tenantID)
	if err != nil {
		return uuid.Nil, mapPgError(err, "invalid or expired reset token")
	}
	return tenantID, nil
}

func (s *Store) MarkTenantPasswordResetTokenUsed(ctx context.Context, tokenHash string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE tenant_password_reset_tokens
		SET used_at = NOW()
		WHERE token_hash = $1 AND used_at IS NULL
	`, tokenHash)
	if err != nil {
		return mapPgError(err, "")
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("invalid or expired reset token")
	}
	return nil
}
