package group

import (
	"context"
	"time"
	"video-conference-be/internal/domain/user"
)

type Group struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	Name        string      `json:"name" gorm:"uniqueIndex"`
	Description string      `json:"description"`
	Members     []user.User `json:"members" gorm:"many2many:group_members;"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type GroupMember struct {
	GroupID  uint      `gorm:"primaryKey"`
	UserID   uint      `gorm:"primaryKey"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type Repository interface {
	Create(ctx context.Context, g *Group) error
	FindByID(ctx context.Context, id uint) (*Group, error)
	List(ctx context.Context) ([]Group, error)
	Update(ctx context.Context, g *Group) error
	Delete(ctx context.Context, id uint) error
	AddMember(ctx context.Context, groupID, userID uint) error
	RemoveMember(ctx context.Context, groupID, userID uint) error

	IsMember(ctx context.Context, groupID, userID uint) (bool, error)
}
