package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/pkg/apperror"
)

func (s *Store) GetSettings(ctx context.Context, orgID uuid.UUID) (*domain.Organization, error) {
	var org domain.Organization
	err := s.pool.QueryRow(ctx, `
		SELECT id, name, created_at, updated_at
		FROM organizations WHERE id = $1
	`, orgID).Scan(&org.ID, &org.Name, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		return nil, mapPgError(err, "organization not found")
	}
	return &org, nil
}

func (s *Store) UpdateOrgName(ctx context.Context, orgID uuid.UUID, name string) (*domain.Organization, error) {
	var org domain.Organization
	err := s.pool.QueryRow(ctx, `
		UPDATE organizations SET name = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, name, created_at, updated_at
	`, name, orgID).Scan(&org.ID, &org.Name, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		return nil, mapPgError(err, "organization not found")
	}
	return &org, nil
}

func (s *Store) CreateStaff(ctx context.Context, staff *domain.Staff) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO staff (id, organization_id, email, password_hash, full_name, is_owner)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, staff.ID, staff.OrganizationID, staff.Email, staff.PasswordHash, staff.FullName, staff.IsOwner)
	return mapPgError(err, "")
}

func (s *Store) GetStaffByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Staff, error) {
	var staff domain.Staff
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, email, password_hash, full_name, is_owner, created_at
		FROM staff WHERE id = $1 AND organization_id = $2
	`, id, orgID).Scan(&staff.ID, &staff.OrganizationID, &staff.Email, &staff.PasswordHash, &staff.FullName, &staff.IsOwner, &staff.CreatedAt)
	if err != nil {
		return nil, mapPgError(err, "staff not found")
	}
	return &staff, nil
}

func (s *Store) GetStaffByEmailAndOrg(ctx context.Context, orgID uuid.UUID, email string) (*domain.Staff, error) {
	var staff domain.Staff
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, email, password_hash, full_name, is_owner, created_at
		FROM staff WHERE organization_id = $1 AND email = $2
	`, orgID, email).Scan(&staff.ID, &staff.OrganizationID, &staff.Email, &staff.PasswordHash, &staff.FullName, &staff.IsOwner, &staff.CreatedAt)
	if err != nil {
		return nil, mapPgError(err, "staff not found")
	}
	return &staff, nil
}

func (s *Store) ListStaff(ctx context.Context, orgID uuid.UUID) ([]domain.StaffProfile, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, email, full_name, is_owner, created_at
		FROM staff WHERE organization_id = $1 ORDER BY created_at
	`, orgID)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()

	var list []domain.StaffProfile
	for rows.Next() {
		var p domain.StaffProfile
		if err := rows.Scan(&p.ID, &p.OrganizationID, &p.Email, &p.FullName, &p.IsOwner, &p.CreatedAt); err != nil {
			return nil, apperror.Internal("scan staff", err)
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (s *Store) DeleteStaff(ctx context.Context, orgID, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `
		DELETE FROM staff WHERE id = $1 AND organization_id = $2 AND is_owner = false
	`, id, orgID)
	if err != nil {
		return mapPgError(err, "")
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("staff not found")
	}
	return nil
}
