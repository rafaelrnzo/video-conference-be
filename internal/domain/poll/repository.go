package poll

import "context"

type Repository interface {
	Create(ctx context.Context, p *Poll) error
}
