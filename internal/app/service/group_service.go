package service

import (
	"context"
	"errors"
	"video-conference-be/internal/domain/group"
	"video-conference-be/internal/domain/user"
)

type GroupService interface {
	ListGroups(ctx context.Context) ([]group.Group, error)
	GetGroup(ctx context.Context, id uint) (*group.Group, error)
	CreateGroup(ctx context.Context, name, description string) (*group.Group, error)
	UpdateGroup(ctx context.Context, id uint, name, description string) (*group.Group, error)
	DeleteGroup(ctx context.Context, id uint) error
	AddMember(ctx context.Context, groupID, userID uint) error
	RemoveMember(ctx context.Context, groupID, userID uint) error

	// FUNGSI BARU
	IsMember(ctx context.Context, groupID, userID uint) (bool, error)
}

type groupService struct {
	groupRepo group.Repository
	userRepo  user.Repository
}

func NewGroupService(gRepo group.Repository, uRepo user.Repository) GroupService {
	return &groupService{
		groupRepo: gRepo,
		userRepo:  uRepo,
	}
}

func (s *groupService) ListGroups(ctx context.Context) ([]group.Group, error) {
	return s.groupRepo.List(ctx)
}

func (s *groupService) GetGroup(ctx context.Context, id uint) (*group.Group, error) {
	return s.groupRepo.FindByID(ctx, id)
}

func (s *groupService) CreateGroup(ctx context.Context, name, description string) (*group.Group, error) {
	if name == "" {
		return nil, errors.New("group name is required")
	}

	g := &group.Group{
		Name:        name,
		Description: description,
	}

	if err := s.groupRepo.Create(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *groupService) UpdateGroup(ctx context.Context, id uint, name, description string) (*group.Group, error) {
	g, err := s.groupRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		g.Name = name
	}
	g.Description = description

	if err := s.groupRepo.Update(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *groupService) DeleteGroup(ctx context.Context, id uint) error {
	if _, err := s.groupRepo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.groupRepo.Delete(ctx, id)
}

func (s *groupService) AddMember(ctx context.Context, groupID, userID uint) error {
	if _, err := s.userRepo.FindByID(ctx, userID); err != nil {
		return errors.New("user not found")
	}
	if _, err := s.groupRepo.FindByID(ctx, groupID); err != nil {
		return errors.New("group not found")
	}
	return s.groupRepo.AddMember(ctx, groupID, userID)
}

func (s *groupService) RemoveMember(ctx context.Context, groupID, userID uint) error {
	if _, err := s.groupRepo.FindByID(ctx, groupID); err != nil {
		return errors.New("group not found")
	}
	return s.groupRepo.RemoveMember(ctx, groupID, userID)
}

// IMPLEMENTASI SERVICE
func (s *groupService) IsMember(ctx context.Context, groupID, userID uint) (bool, error) {
	return s.groupRepo.IsMember(ctx, groupID, userID)
}
