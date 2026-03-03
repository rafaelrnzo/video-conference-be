package service

import (
	"context"
	"time"

	dRecord "video-conference-be/internal/domain/record"
)

type RecordService interface {
	Create(ctx context.Context, roomID, name, link, egressID, status string) (*dRecord.Record, error)
	ListAll(ctx context.Context) ([]*dRecord.Record, error)
	ListByRoomID(ctx context.Context, roomID string) ([]*dRecord.Record, error)
	GetByID(ctx context.Context, id uint) (*dRecord.Record, error)
	UpdateName(ctx context.Context, id uint, newName string) (*dRecord.Record, error)
	UpdateStatus(ctx context.Context, id uint, status string) (*dRecord.Record, error)
	Delete(ctx context.Context, id uint) error
	Exists(ctx context.Context, link string) (bool, error)
}

type recordService struct {
	repo dRecord.Repository
}

func NewRecordService(repo dRecord.Repository) RecordService {
	return &recordService{repo: repo}
}

func (s *recordService) Create(ctx context.Context, roomID, name, link, egressID, status string) (*dRecord.Record, error) {
	rec := &dRecord.Record{
		RoomID:    roomID,
		Name:      name,
		Link:      link,
		EgressID:  egressID,
		Status:    status,
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(ctx, rec); err != nil {
		return nil, err
	}
	return rec, nil
}

func (s *recordService) ListAll(ctx context.Context) ([]*dRecord.Record, error) {
	return s.repo.ListAll(ctx)
}

func (s *recordService) ListByRoomID(ctx context.Context, roomID string) ([]*dRecord.Record, error) {
	return s.repo.ListByRoomID(ctx, roomID)
}

func (s *recordService) GetByID(ctx context.Context, id uint) (*dRecord.Record, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *recordService) UpdateName(ctx context.Context, id uint, newName string) (*dRecord.Record, error) {
	rec, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	rec.Name = newName
	if err := s.repo.Update(ctx, rec); err != nil {
		return nil, err
	}
	return rec, nil
}

func (s *recordService) UpdateStatus(ctx context.Context, id uint, status string) (*dRecord.Record, error) {
	rec, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	rec.Status = status
	if err := s.repo.Update(ctx, rec); err != nil {
		return nil, err
	}
	return rec, nil
}

func (s *recordService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *recordService) Exists(ctx context.Context, link string) (bool, error) {
	return s.repo.Exists(ctx, link)
}
