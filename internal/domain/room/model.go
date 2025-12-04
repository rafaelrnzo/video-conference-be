package room

import "time"

type Room struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	Name            string    `json:"name" gorm:"uniqueIndex"`
	RoomCode        string    `json:"room_code" gorm:"uniqueIndex"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Description     string    `json:"description"` // Field ini wajib ada
	MaxParticipants int       `json:"max_participants"`
	CreatedByID     uint      `json:"createdById"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
