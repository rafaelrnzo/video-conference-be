package repository

import (
	"context"

	"video-conference-be/internal/domain/role"
	"video-conference-be/pkg/utility"

	"gorm.io/gorm"
)

type roleGormRepo struct {
	db *gorm.DB
}

func NewRoleRepository() role.Repository {
	return &roleGormRepo{db: utility.DB}
}

func (r *roleGormRepo) Create(ctx context.Context, role *role.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleGormRepo) Update(ctx context.Context, role *role.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleGormRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&role.Role{}, id).Error
}

func (r *roleGormRepo) FindByID(ctx context.Context, id uint) (*role.Role, error) {
	var ro role.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").First(&ro, id).Error; err != nil {
		return nil, err
	}
	return &ro, nil
}

func (r *roleGormRepo) FindByName(ctx context.Context, name string) (*role.Role, error) {
	var ro role.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").Where("name = ?", name).First(&ro).Error; err != nil {
		return nil, err
	}
	return &ro, nil
}

func (r *roleGormRepo) List(ctx context.Context) ([]role.Role, error) {
	var roles []role.Role
	if err := r.db.WithContext(ctx).Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleGormRepo) AssignPermission(ctx context.Context, roleID, permID uint) error {
	var ro role.Role
	if err := r.db.First(&ro, roleID).Error; err != nil {
		return err
	}
	var perm role.Permission
	if err := r.db.First(&perm, permID).Error; err != nil {
		return err
	}
	return r.db.Model(&ro).Association("Permissions").Append(&perm)
}

func (r *roleGormRepo) RevokePermission(ctx context.Context, roleID, permID uint) error {
	var ro role.Role
	if err := r.db.First(&ro, roleID).Error; err != nil {
		return err
	}
	var perm role.Permission
	if err := r.db.First(&perm, permID).Error; err != nil {
		return err
	}
	return r.db.Model(&ro).Association("Permissions").Delete(&perm)
}

func (r *roleGormRepo) ListPermissions(ctx context.Context) ([]role.Permission, error) {
    var perms []role.Permission
    if err := r.db.WithContext(ctx).Find(&perms).Error; err != nil {
        return nil, err
    }
    return perms, nil
}

func (r *roleGormRepo) CreatePermission(ctx context.Context, perm *role.Permission) error {
    return r.db.WithContext(ctx).Create(perm).Error
}

func (r *roleGormRepo) FindPermissionByKey(ctx context.Context, key string) (*role.Permission, error) {
    var perm role.Permission
    if err := r.db.WithContext(ctx).Where("key = ?", key).First(&perm).Error; err != nil {
        return nil, err
    }
    return &perm, nil
}
