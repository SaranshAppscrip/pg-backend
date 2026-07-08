package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/pkg/apperror"
)

func (s *Store) ListProperties(ctx context.Context, orgID uuid.UUID) ([]domain.Property, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, name, address, created_at, updated_at
		FROM properties WHERE organization_id = $1 ORDER BY name
	`, orgID)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()

	var list []domain.Property
	for rows.Next() {
		var p domain.Property
		if err := rows.Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Address, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, apperror.Internal("scan property", err)
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

func (s *Store) GetPropertyByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Property, error) {
	var p domain.Property
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, name, address, created_at, updated_at
		FROM properties WHERE id = $1 AND organization_id = $2
	`, id, orgID).Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Address, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, mapPgError(err, "property not found")
	}
	return &p, nil
}

func (s *Store) CreateProperty(ctx context.Context, property *domain.Property) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO properties (id, organization_id, name, address)
		VALUES ($1, $2, $3, $4)
	`, property.ID, property.OrganizationID, property.Name, property.Address)
	return mapPgError(err, "")
}

func (s *Store) UpdateProperty(ctx context.Context, orgID, id uuid.UUID, name string, address *string) (*domain.Property, error) {
	var p domain.Property
	err := s.pool.QueryRow(ctx, `
		UPDATE properties SET name = $3, address = $4, updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
		RETURNING id, organization_id, name, address, created_at, updated_at
	`, id, orgID, name, address).Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Address, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, mapPgError(err, "property not found")
	}
	return &p, nil
}
