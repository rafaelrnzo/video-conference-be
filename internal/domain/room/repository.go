package room

import "context"

type Repository interface {
	Create(ctx context.Context, r *Room) error
	ListAll(ctx context.Context) ([]*Room, error)
	GetByName(ctx context.Context, name string) (*Room, error)
	Delete(ctx context.Context, name string) error
}
