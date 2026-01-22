package user

import (
	"time"
	"video-conference-be/internal/domain/role"
)

type User struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`

	// Foreign Key to Role
	RoleID uint      `json:"role_id"`
	Role   role.Role `json:"role" gorm:"foreignKey:RoleID"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
