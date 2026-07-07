package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/pkg/apperror"
)

func (s *Store) ListTenants(ctx context.Context, orgID uuid.UUID) ([]domain.Tenant, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, name, email, phone, room_id, monthly_fee, join_date, active, created_at
		FROM tenants WHERE organization_id = $1 ORDER BY created_at
	`, orgID)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()

	var list []domain.Tenant
	for rows.Next() {
		var t domain.Tenant
		if err := rows.Scan(&t.ID, &t.OrganizationID, &t.Name, &t.Email, &t.Phone, &t.RoomID, &t.MonthlyFee, &t.JoinDate, &t.Active, &t.CreatedAt); err != nil {
			return nil, apperror.Internal("scan tenant", err)
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (s *Store) CreateTenant(ctx context.Context, tenant *domain.Tenant) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO tenants (id, organization_id, name, email, password_hash, phone, room_id, monthly_fee, join_date, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, tenant.ID, tenant.OrganizationID, tenant.Name, tenant.Email, tenant.PasswordHash, tenant.Phone, tenant.RoomID, tenant.MonthlyFee, tenant.JoinDate, tenant.Active)
	return mapPgError(err, "")
}

func (s *Store) GetTenantByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Tenant, error) {
	var t domain.Tenant
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, name, email, phone, room_id, monthly_fee, join_date, active, created_at
		FROM tenants WHERE id = $1 AND organization_id = $2
	`, id, orgID).Scan(&t.ID, &t.OrganizationID, &t.Name, &t.Email, &t.Phone, &t.RoomID, &t.MonthlyFee, &t.JoinDate, &t.Active, &t.CreatedAt)
	if err != nil {
		return nil, mapPgError(err, "tenant not found")
	}
	return &t, nil
}

func (s *Store) GetTenantByEmailAndOrg(ctx context.Context, orgID uuid.UUID, email string) (*domain.Tenant, error) {
	var t domain.Tenant
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, name, email, password_hash, phone, room_id, monthly_fee, join_date, active, created_at
		FROM tenants WHERE organization_id = $1 AND email = $2
	`, orgID, email).Scan(&t.ID, &t.OrganizationID, &t.Name, &t.Email, &t.PasswordHash, &t.Phone, &t.RoomID, &t.MonthlyFee, &t.JoinDate, &t.Active, &t.CreatedAt)
	if err != nil {
		return nil, mapPgError(err, "tenant not found")
	}
	return &t, nil
}

func (s *Store) ListTenantsByEmail(ctx context.Context, email string) ([]domain.Tenant, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, name, email, password_hash, phone, room_id, monthly_fee, join_date, active, created_at
		FROM tenants WHERE lower(trim(email)) = lower(trim($1))
	`, email)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()

	var list []domain.Tenant
	for rows.Next() {
		var t domain.Tenant
		if err := rows.Scan(&t.ID, &t.OrganizationID, &t.Name, &t.Email, &t.PasswordHash, &t.Phone, &t.RoomID, &t.MonthlyFee, &t.JoinDate, &t.Active, &t.CreatedAt); err != nil {
			return nil, apperror.Internal("scan tenant", err)
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (s *Store) GetTenantByPhoneAndOrg(ctx context.Context, orgID uuid.UUID, phone string) (*domain.Tenant, error) {
	var t domain.Tenant
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, name, email, phone, room_id, monthly_fee, join_date, active, created_at
		FROM tenants WHERE organization_id = $1 AND phone = $2 AND active = true
	`, orgID, phone).Scan(&t.ID, &t.OrganizationID, &t.Name, &t.Email, &t.Phone, &t.RoomID, &t.MonthlyFee, &t.JoinDate, &t.Active, &t.CreatedAt)
	if err != nil {
		return nil, mapPgError(err, "tenant not found")
	}
	return &t, nil
}

func (s *Store) GetTenantByNameAndOrg(ctx context.Context, orgID uuid.UUID, name string) (*domain.Tenant, error) {
	var t domain.Tenant
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, name, email, phone, room_id, monthly_fee, join_date, active, created_at
		FROM tenants
		WHERE organization_id = $1 AND lower(trim(name)) = lower(trim($2)) AND active = true
	`, orgID, name).Scan(&t.ID, &t.OrganizationID, &t.Name, &t.Email, &t.Phone, &t.RoomID, &t.MonthlyFee, &t.JoinDate, &t.Active, &t.CreatedAt)
	if err != nil {
		return nil, mapPgError(err, "tenant not found")
	}
	return &t, nil
}

func (s *Store) MoveOutTenant(ctx context.Context, orgID, id uuid.UUID) (*domain.Tenant, error) {
	var t domain.Tenant
	err := s.pool.QueryRow(ctx, `
		UPDATE tenants SET active = false WHERE id = $1 AND organization_id = $2
		RETURNING id, organization_id, name, email, phone, room_id, monthly_fee, join_date, active, created_at
	`, id, orgID).Scan(&t.ID, &t.OrganizationID, &t.Name, &t.Email, &t.Phone, &t.RoomID, &t.MonthlyFee, &t.JoinDate, &t.Active, &t.CreatedAt)
	if err != nil {
		return nil, mapPgError(err, "tenant not found")
	}
	return &t, nil
}

func (s *Store) GetTenantProfile(ctx context.Context, orgID, id uuid.UUID) (*domain.TenantProfile, error) {
	var p domain.TenantProfile
	var roomNumber *string
	err := s.pool.QueryRow(ctx, `
		SELECT t.id, t.organization_id, t.name, t.email, t.phone, t.monthly_fee, t.join_date, r.room_number
		FROM tenants t
		LEFT JOIN rooms r ON r.id = t.room_id AND r.organization_id = t.organization_id
		WHERE t.id = $1 AND t.organization_id = $2 AND t.active = true
	`, id, orgID).Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Email, &p.Phone, &p.MonthlyFee, &p.JoinDate, &roomNumber)
	if err != nil {
		return nil, mapPgError(err, "tenant not found")
	}
	p.RoomNumber = roomNumber
	return &p, nil
}
