package record

import "context"

type Repository interface {
	Create(ctx context.Context, r *Record) error
	ListAll(ctx context.Context) ([]*Record, error)
	ListByRoomID(ctx context.Context, roomID string) ([]*Record, error)
	GetByID(ctx context.Context, id uint) (*Record, error)
	Update(ctx context.Context, r *Record) error
	Delete(ctx context.Context, id uint) error
}
