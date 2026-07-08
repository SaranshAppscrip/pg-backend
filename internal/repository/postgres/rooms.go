package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/pkg/apperror"
)

func (s *Store) ListRooms(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]domain.Room, error) {
	query := `
		SELECT id, organization_id, property_id, room_number, capacity, created_at
		FROM rooms WHERE organization_id = $1`
	args := []any{orgID}
	if propertyID != nil {
		query += ` AND property_id = $2`
		args = append(args, *propertyID)
	}
	query += ` ORDER BY room_number`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()

	var list []domain.Room
	for rows.Next() {
		var r domain.Room
		if err := rows.Scan(&r.ID, &r.OrganizationID, &r.PropertyID, &r.RoomNumber, &r.Capacity, &r.CreatedAt); err != nil {
			return nil, apperror.Internal("scan room", err)
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

func (s *Store) CreateRoom(ctx context.Context, room *domain.Room) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO rooms (id, organization_id, property_id, room_number, capacity)
		VALUES ($1, $2, $3, $4, $5)
	`, room.ID, room.OrganizationID, room.PropertyID, room.RoomNumber, room.Capacity)
	return mapPgError(err, "")
}

func (s *Store) GetRoomByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Room, error) {
	var r domain.Room
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, property_id, room_number, capacity, created_at
		FROM rooms WHERE id = $1 AND organization_id = $2
	`, id, orgID).Scan(&r.ID, &r.OrganizationID, &r.PropertyID, &r.RoomNumber, &r.Capacity, &r.CreatedAt)
	if err != nil {
		return nil, mapPgError(err, "room not found")
	}
	return &r, nil
}

func (s *Store) GetRoomByNumber(ctx context.Context, propertyID uuid.UUID, roomNumber string) (*domain.Room, error) {
	var r domain.Room
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, property_id, room_number, capacity, created_at
		FROM rooms WHERE property_id = $1 AND room_number = $2
	`, propertyID, roomNumber).Scan(&r.ID, &r.OrganizationID, &r.PropertyID, &r.RoomNumber, &r.Capacity, &r.CreatedAt)
	if err != nil {
		return nil, mapPgError(err, "room not found")
	}
	return &r, nil
}

func (s *Store) DeleteRoom(ctx context.Context, orgID, id uuid.UUID) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM rooms WHERE id = $1 AND organization_id = $2`, id, orgID)
	if err != nil {
		return mapPgError(err, "")
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("room not found")
	}
	return nil
}

func (s *Store) CountActiveTenantsInRoom(ctx context.Context, orgID, roomID uuid.UUID) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM tenants
		WHERE organization_id = $1 AND room_id = $2 AND active = true
	`, orgID, roomID).Scan(&count)
	return count, mapPgError(err, "")
}
