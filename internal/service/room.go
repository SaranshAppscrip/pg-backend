package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
)

type RoomService struct {
	repos repository.Store
}

func NewRoomService(repos repository.Store) *RoomService {
	return &RoomService{repos: repos}
}

func (s *RoomService) List(ctx context.Context, orgID uuid.UUID) ([]domain.Room, error) {
	return s.repos.Rooms.List(ctx, orgID)
}

func (s *RoomService) Create(ctx context.Context, orgID uuid.UUID, roomNumber string, capacity int) (*domain.Room, error) {
	if roomNumber == "" || capacity < 1 {
		return nil, apperror.BadRequest("room_number and capacity (>= 1) are required")
	}
	room := &domain.Room{
		ID:             uuid.New(),
		OrganizationID: orgID,
		RoomNumber:     roomNumber,
		Capacity:       capacity,
		CreatedAt:      time.Now(),
	}
	if err := s.repos.Rooms.Create(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) Delete(ctx context.Context, orgID, id uuid.UUID) error {
	count, err := s.repos.Rooms.CountActiveTenants(ctx, orgID, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return apperror.Conflict("cannot remove room with active tenants")
	}
	return s.repos.Rooms.Delete(ctx, orgID, id)
}
