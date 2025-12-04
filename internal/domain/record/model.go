package record

import "time"

type Record struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	RoomID    string    `json:"room_id"`
	Name      string    `json:"name"`
	Link      string    `json:"link"`
	EgressID  string    `json:"egress_id"`
	CreatedAt time.Time `json:"created_at"`
}
