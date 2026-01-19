package repository

import (
	"context"

	"video-conference-be/internal/domain/poll"
	"video-conference-be/pkg/utility"

	"gorm.io/gorm"
)

type pollGormRepo struct {
	db *gorm.DB
}

func NewPollRepository() poll.Repository {
	return &pollGormRepo{db: utility.DB}
}

func (r *pollGormRepo) Create(ctx context.Context, p *poll.Poll) error {
	return r.db.WithContext(ctx).Create(p).Error
}
