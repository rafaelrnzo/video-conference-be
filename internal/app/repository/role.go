package repository

import (
	"video-conference-be/internal/domain/role"
	"video-conference-be/pkg/utility"

	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(r *role.Role) error
	FindAll() ([]role.Role, error)
	FindByName(name string) (*role.Role, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository() RoleRepository {
	return &roleRepository{
		db: utility.DB,
	}
}

func (r *roleRepository) Create(newRole *role.Role) error {
	return r.db.Create(newRole).Error
}

func (r *roleRepository) FindAll() ([]role.Role, error) {
	var roles []role.Role
	err := r.db.Find(&roles).Error
	return roles, err
}

func (r *roleRepository) FindByName(name string) (*role.Role, error) {
	var role role.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	return &role, err
}
