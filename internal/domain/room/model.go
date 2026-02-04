package room

import (
	"time"

	"video-conference-be/internal/domain/group"

	"github.com/lib/pq"
)

type Room struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Name            string         `json:"name" gorm:"uniqueIndex"`
	RoomCode        string         `json:"room_code" gorm:"uniqueIndex"`
	StartDate       time.Time      `json:"start_date"`
	EndDate         time.Time      `json:"end_date"`
	GroupID         *uint          `json:"group_id"`
	Group           *group.Group   `json:"group" gorm:"foreignKey:GroupID"`
	AssignedTo      pq.StringArray `json:"assigned_to" gorm:"type:text[]"`
	BannedUsers     pq.StringArray `json:"banned_users" gorm:"type:text[]"`
	Description     string         `json:"description"`
	MaxParticipants int            `json:"max_participants"`
	PresentationPath string        `json:"presentation_path"`
	CreatedByID     uint           `json:"createdById"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}
