package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
)

type KitchenService struct {
	repos repository.KitchenRepository
}

func NewKitchenService(repos repository.KitchenRepository) *KitchenService {
	return &KitchenService{repos: repos}
}

func (s *KitchenService) ListItems(ctx context.Context, orgID uuid.UUID) ([]domain.KitchenItem, error) {
	return s.repos.ListItems(ctx, orgID)
}

type CreateKitchenItemInput struct {
	Name             string
	Qty              float64
	Unit             domain.KitchenUnit
	ReorderThreshold float64
}

func (s *KitchenService) CreateItem(ctx context.Context, orgID uuid.UUID, in CreateKitchenItemInput) (*domain.KitchenItem, error) {
	if in.Name == "" {
		return nil, apperror.BadRequest("name is required")
	}
	item := &domain.KitchenItem{
		ID:               uuid.New(),
		OrganizationID:   orgID,
		Name:             in.Name,
		Qty:              in.Qty,
		Unit:             in.Unit,
		ReorderThreshold: in.ReorderThreshold,
		CreatedAt:        time.Now(),
	}
	if err := s.repos.CreateItem(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

type StockMovementInput struct {
	Qty  float64
	Date time.Time
	Note *string
}

func (s *KitchenService) StockIn(ctx context.Context, orgID, id uuid.UUID, in StockMovementInput) (*domain.KitchenItem, error) {
	item, err := s.repos.GetItem(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	newQty := item.Qty + in.Qty
	if err := s.repos.UpdateItemQty(ctx, orgID, id, newQty); err != nil {
		return nil, err
	}
	log := &domain.KitchenLog{
		ID:     uuid.New(),
		ItemID: id,
		Type:   domain.KitchenLogIn,
		Qty:    in.Qty,
		Date:   in.Date,
		Note:   in.Note,
	}
	if err := s.repos.CreateLog(ctx, orgID, log); err != nil {
		return nil, err
	}
	item.Qty = newQty
	return item, nil
}

func (s *KitchenService) UseStock(ctx context.Context, orgID, id uuid.UUID, in StockMovementInput) (*domain.KitchenItem, error) {
	item, err := s.repos.GetItem(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	if in.Qty > item.Qty {
		return nil, apperror.BadRequest("cannot use more than current stock")
	}
	newQty := item.Qty - in.Qty
	if err := s.repos.UpdateItemQty(ctx, orgID, id, newQty); err != nil {
		return nil, err
	}
	log := &domain.KitchenLog{
		ID:     uuid.New(),
		ItemID: id,
		Type:   domain.KitchenLogOut,
		Qty:    in.Qty,
		Date:   in.Date,
		Note:   in.Note,
	}
	if err := s.repos.CreateLog(ctx, orgID, log); err != nil {
		return nil, err
	}
	item.Qty = newQty
	return item, nil
}

func (s *KitchenService) ListLog(ctx context.Context, orgID uuid.UUID, limit int) ([]domain.KitchenLog, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repos.ListLog(ctx, orgID, limit)
}
