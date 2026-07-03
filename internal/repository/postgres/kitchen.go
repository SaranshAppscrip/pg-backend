package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/pkg/apperror"
)

func (s *Store) ListKitchenItems(ctx context.Context, orgID uuid.UUID) ([]domain.KitchenItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, organization_id, name, qty, unit, reorder_threshold, created_at
		FROM kitchen_items WHERE organization_id = $1 ORDER BY name
	`, orgID)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()

	var list []domain.KitchenItem
	for rows.Next() {
		var item domain.KitchenItem
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.Name, &item.Qty, &item.Unit, &item.ReorderThreshold, &item.CreatedAt); err != nil {
			return nil, apperror.Internal("scan kitchen item", err)
		}
		list = append(list, item)
	}
	return list, rows.Err()
}

func (s *Store) CreateKitchenItem(ctx context.Context, item *domain.KitchenItem) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO kitchen_items (id, organization_id, name, qty, unit, reorder_threshold)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, item.ID, item.OrganizationID, item.Name, item.Qty, item.Unit, item.ReorderThreshold)
	return mapPgError(err, "")
}

func (s *Store) GetKitchenItem(ctx context.Context, orgID, id uuid.UUID) (*domain.KitchenItem, error) {
	var item domain.KitchenItem
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, name, qty, unit, reorder_threshold, created_at
		FROM kitchen_items WHERE id = $1 AND organization_id = $2
	`, id, orgID).Scan(&item.ID, &item.OrganizationID, &item.Name, &item.Qty, &item.Unit, &item.ReorderThreshold, &item.CreatedAt)
	if err != nil {
		return nil, mapPgError(err, "kitchen item not found")
	}
	return &item, nil
}

func (s *Store) GetKitchenItemByName(ctx context.Context, orgID uuid.UUID, name string) (*domain.KitchenItem, error) {
	var item domain.KitchenItem
	err := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, name, qty, unit, reorder_threshold, created_at
		FROM kitchen_items
		WHERE organization_id = $1 AND lower(trim(name)) = lower(trim($2))
	`, orgID, name).Scan(&item.ID, &item.OrganizationID, &item.Name, &item.Qty, &item.Unit, &item.ReorderThreshold, &item.CreatedAt)
	if err != nil {
		return nil, mapPgError(err, "kitchen item not found")
	}
	return &item, nil
}

func (s *Store) UpdateKitchenItemQty(ctx context.Context, orgID, id uuid.UUID, qty float64) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE kitchen_items SET qty = $1 WHERE id = $2 AND organization_id = $3
	`, qty, id, orgID)
	if err != nil {
		return mapPgError(err, "")
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("kitchen item not found")
	}
	return nil
}

func (s *Store) CreateKitchenLog(ctx context.Context, orgID uuid.UUID, log *domain.KitchenLog) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO kitchen_log (id, organization_id, item_id, type, qty, date, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, log.ID, orgID, log.ItemID, log.Type, log.Qty, log.Date, log.Note)
	return mapPgError(err, "")
}

func (s *Store) ListKitchenLog(ctx context.Context, orgID uuid.UUID, limit int) ([]domain.KitchenLog, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, item_id, type, qty, date, note, created_at
		FROM kitchen_log WHERE organization_id = $1 ORDER BY date DESC, created_at DESC LIMIT $2
	`, orgID, limit)
	if err != nil {
		return nil, mapPgError(err, "")
	}
	defer rows.Close()

	var list []domain.KitchenLog
	for rows.Next() {
		var l domain.KitchenLog
		if err := rows.Scan(&l.ID, &l.ItemID, &l.Type, &l.Qty, &l.Date, &l.Note, &l.CreatedAt); err != nil {
			return nil, apperror.Internal("scan kitchen log", err)
		}
		list = append(list, l)
	}
	return list, rows.Err()
}
