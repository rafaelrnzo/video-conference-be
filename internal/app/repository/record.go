package repository

import (
	"context"

	dRecord "video-conference-be/internal/domain/record"
	"video-conference-be/pkg/utility"

	"gorm.io/gorm"
)

type recordRepository struct {
	db *gorm.DB
}

func NewRecordRepository() dRecord.Repository {
	return &recordRepository{
		db: utility.DB,
	}
}

func (r *recordRepository) Create(ctx context.Context, rec *dRecord.Record) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *recordRepository) ListAll(ctx context.Context) ([]*dRecord.Record, error) {
	var out []*dRecord.Record
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&out).Error
	return out, err
}

func (r *recordRepository) ListByRoomID(ctx context.Context, roomID string) ([]*dRecord.Record, error) {
	var out []*dRecord.Record
	err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		Find(&out).Error
	return out, err
}

func (r *recordRepository) GetByID(ctx context.Context, id uint) (*dRecord.Record, error) {
	var rec dRecord.Record
	err := r.db.WithContext(ctx).First(&rec, id).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *recordRepository) Update(ctx context.Context, rec *dRecord.Record) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *recordRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&dRecord.Record{}, id).Error
}

func (r *recordRepository) Exists(ctx context.Context, link string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&dRecord.Record{}).Where("link = ?", link).Count(&count).Error
	return count > 0, err
}
