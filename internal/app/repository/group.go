package repository

import (
	"context"
	"time"

	"video-conference-be/internal/domain/group"
	"video-conference-be/internal/domain/user"
	"video-conference-be/pkg/utility"

	"gorm.io/gorm"
)

type groupGormRepo struct {
	db *gorm.DB
}

func NewGroupRepository() group.Repository {
	return &groupGormRepo{db: utility.DB}
}

func (r *groupGormRepo) Create(ctx context.Context, g *group.Group) error {
	return r.db.WithContext(ctx).Create(g).Error
}

func (r *groupGormRepo) FindByID(ctx context.Context, id uint) (*group.Group, error) {
	var g group.Group
	if err := r.db.WithContext(ctx).Preload("Members").First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *groupGormRepo) List(ctx context.Context) ([]group.Group, error) {
	var groups []group.Group
	if err := r.db.WithContext(ctx).Preload("Members").Order("created_at DESC").Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *groupGormRepo) Update(ctx context.Context, g *group.Group) error {
	return r.db.WithContext(ctx).Save(g).Error
}

func (r *groupGormRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&group.Group{}, id).Error
}

func (r *groupGormRepo) AddMember(ctx context.Context, groupID, userID uint) error {
	var g group.Group
	g.ID = groupID
	var u user.User
	u.ID = userID

	if err := r.db.WithContext(ctx).Model(&g).Association("Members").Append(&u); err != nil {
		return err
	}

	r.db.WithContext(ctx).Model(&group.GroupMember{}).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Update("joined_at", time.Now())

	return nil
}

func (r *groupGormRepo) RemoveMember(ctx context.Context, groupID, userID uint) error {
	var g group.Group
	g.ID = groupID
	var u user.User
	u.ID = userID

	return r.db.WithContext(ctx).Model(&g).Association("Members").Delete(&u)
}

func (r *groupGormRepo) IsMember(ctx context.Context, groupID, userID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("group_members").
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}
