package service

import (
	"errors"
	"video-conference-be/internal/domain/room"
	"video-conference-be/pkg/utility"

	"gorm.io/gorm"
)

type RoomService interface {
	ListRooms(userID uint, username, role string) ([]room.Room, error)
	GetRoomByCode(code string) (*room.Room, error)
	CreateRoom(req room.Room) (*room.Room, error)
	UpdateRoom(req room.Room) (*room.Room, error)
	DeleteRoom(id uint) error
}

type roomService struct{}

func NewRoomService() RoomService {
	return &roomService{}
}

func (s *roomService) ListRooms(userID uint, username, role string) ([]room.Room, error) {
	var rooms []room.Room

	query := utility.DB.
		Select("id", "name", "room_code", "group_id", "start_date", "end_date",
			"description", "max_participants", "assigned_to",
			"created_by_id", "created_at", "updated_at").
		Preload("Group").
		Order("created_at ASC")

	if role != "admin" {
		query = query.Where(`
			(group_id IS NULL) OR 
			(group_id IN (SELECT group_id FROM group_members WHERE user_id = ?)) OR
			(? = ANY(assigned_to))
		`, userID, username)
	}

	if err := query.Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (s *roomService) GetRoomByCode(code string) (*room.Room, error) {
	var r room.Room
	if err := utility.DB.Preload("Group").
		Where("room_code = ?", code).
		First(&r).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("room not found")
		}
		return nil, err
	}
	return &r, nil
}

func (s *roomService) CreateRoom(req room.Room) (*room.Room, error) {
	if req.Name == "" {
		return nil, errors.New("room name is required")
	}
	if req.StartDate.IsZero() || req.EndDate.IsZero() {
		return nil, errors.New("start date and end date are required")
	}
	if req.EndDate.Before(req.StartDate) {
		return nil, errors.New("end date cannot be before start date")
	}

	if req.GroupID != nil && *req.GroupID == 0 {
		req.GroupID = nil
	}

	req.RoomCode = utility.RandomToken(10)

	if err := utility.DB.Create(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *roomService) UpdateRoom(req room.Room) (*room.Room, error) {
	if req.ID == 0 {
		return nil, errors.New("id is required")
	}

	var existingRoom room.Room
	if err := utility.DB.First(&existingRoom, req.ID).Error; err != nil {
		return nil, errors.New("room not found")
	}

	existingRoom.Name = req.Name
	existingRoom.Description = req.Description
	existingRoom.MaxParticipants = req.MaxParticipants
	existingRoom.AssignedTo = req.AssignedTo

	if req.GroupID != nil {
		if *req.GroupID == 0 {
			existingRoom.GroupID = nil
		} else {
			existingRoom.GroupID = req.GroupID
		}
	}

	if !req.StartDate.IsZero() {
		existingRoom.StartDate = req.StartDate
	}
	if !req.EndDate.IsZero() {
		existingRoom.EndDate = req.EndDate
	}

	if existingRoom.EndDate.Before(existingRoom.StartDate) {
		return nil, errors.New("end date cannot be before start date")
	}

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
