package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/repository"
	"github.com/nivas/server/pkg/apperror"
	"github.com/nivas/server/pkg/logger"
)

type RoomService struct {
	repos repository.Store
}

func NewRoomService(repos repository.Store) *RoomService {
	return &RoomService{repos: repos}
}

func (s *RoomService) List(ctx context.Context, orgID uuid.UUID, propertyID *uuid.UUID) ([]domain.Room, error) {
	return s.repos.Rooms.List(ctx, orgID, propertyID)
}

func (s *RoomService) Create(ctx context.Context, orgID, propertyID uuid.UUID, roomNumber string, capacity int) (*domain.Room, error) {
	log := logger.FromContext(ctx)
	roomNumber = strings.TrimSpace(roomNumber)
	if roomNumber == "" || capacity < 1 {
		return nil, apperror.BadRequest("room_number and capacity (>= 1) are required")
	}

	if _, err := s.repos.Properties.GetByID(ctx, orgID, propertyID); err != nil {
		return nil, err
	}

	if _, err := s.repos.Rooms.GetByRoomNumber(ctx, propertyID, roomNumber); err == nil {
		log.Warn("room create rejected", "organization_id", orgID, "property_id", propertyID, "room_number", roomNumber, "reason", "duplicate_room_number")
		return nil, apperror.DuplicateRoomNumber()
	} else if !apperror.IsNotFound(err) {
		return nil, err
	}

	room := &domain.Room{
		ID:             uuid.New(),
		OrganizationID: orgID,
		PropertyID:     propertyID,
		RoomNumber:     roomNumber,
		Capacity:       capacity,
		CreatedAt:      time.Now(),
	}
	if err := s.repos.Rooms.Create(ctx, room); err != nil {
		return nil, err
	}
	log.Info("room created", "organization_id", orgID, "property_id", propertyID, "room_id", room.ID, "room_number", roomNumber, "capacity", capacity)
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
