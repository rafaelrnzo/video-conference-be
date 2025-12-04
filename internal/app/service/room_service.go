package service

import (
	"errors"
	"video-conference-be/internal/domain/room"
	"video-conference-be/pkg/utility"

	"gorm.io/gorm"
)

type RoomService interface {
	ListRooms() ([]room.Room, error)
	CreateRoom(req room.Room) (*room.Room, error)
	GetRoomByName(name string) (*room.Room, error)
	UpdateRoom(req room.Room) (*room.Room, error)
	DeleteRoom(id uint) error
}

type roomService struct{}

func NewRoomService() RoomService {
	return &roomService{}
}

func (s *roomService) ListRooms() ([]room.Room, error) {
	var rooms []room.Room
	if err := utility.DB.
		Select("id", "name", "description", "max_participants", "created_by_id", "created_at", "updated_at").
		Order("created_at DESC").
		Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (s *roomService) CreateRoom(req room.Room) (*room.Room, error) {
	if req.Name == "" {
		return nil, errors.New("room name is required")
	}
	if req.MaxParticipants <= 0 {
		return nil, errors.New("max participants must be greater than 0")
	}

	if err := utility.DB.Create(&req).Error; err != nil {
		return nil, err
	}

	return &req, nil
}

func (s *roomService) UpdateRoom(req room.Room) (*room.Room, error) {
	var existingRoom room.Room
	if err := utility.DB.First(&existingRoom, req.ID).Error; err != nil {
		return nil, err
	}

	existingRoom.Name = req.Name
	existingRoom.Description = req.Description
	existingRoom.MaxParticipants = req.MaxParticipants

	if err := utility.DB.Save(&existingRoom).Error; err != nil {
		return nil, err
	}

	return &existingRoom, nil
}

func (s *roomService) DeleteRoom(id uint) error {
	if err := utility.DB.Delete(&room.Room{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (s *roomService) GetRoomByName(name string) (*room.Room, error) {
	var r room.Room
	if err := utility.DB.Where("name = ?", name).First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("room not found")
		}
		return nil, err
	}
	return &r, nil
}
