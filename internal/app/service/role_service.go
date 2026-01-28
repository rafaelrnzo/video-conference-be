package service

import (
	"context"

	"video-conference-be/internal/domain/role"
)

type RoleService interface {
	CreateRole(ctx context.Context, name, description string) (*role.Role, error)
	ListRoles(ctx context.Context) ([]role.Role, error)
	UpdateRole(ctx context.Context, id uint, name, description string) (*role.Role, error)
	DeleteRole(ctx context.Context, id uint) error
	
	AssignPermission(ctx context.Context, roleID, permID uint) error
	RevokePermission(ctx context.Context, roleID, permID uint) error
	ListPermissions(ctx context.Context) ([]role.Permission, error)
    
    // Setup defaults
    InitDefaultRoles(ctx context.Context) error
}

type roleService struct {
	repo role.Repository
}

func NewRoleService(repo role.Repository) RoleService {
	return &roleService{repo: repo}
}

func (s *roleService) CreateRole(ctx context.Context, name, description string) (*role.Role, error) {
	r := &role.Role{
		Name:        name,
		Description: description,
	}
	if err := s.repo.Create(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *roleService) ListRoles(ctx context.Context) ([]role.Role, error) {
	return s.repo.List(ctx)
}

func (s *roleService) UpdateRole(ctx context.Context, id uint, name, description string) (*role.Role, error) {
	r, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	r.Name = name
	r.Description = description
	if err := s.repo.Update(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *roleService) DeleteRole(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *roleService) AssignPermission(ctx context.Context, roleID, permID uint) error {
	return s.repo.AssignPermission(ctx, roleID, permID)
}

func (s *roleService) RevokePermission(ctx context.Context, roleID, permID uint) error {
	return s.repo.RevokePermission(ctx, roleID, permID)
}

func (s *roleService) ListPermissions(ctx context.Context) ([]role.Permission, error) {
    return s.repo.ListPermissions(ctx)
}

// InitDefaultRoles ensures basic roles exist
func (s *roleService) InitDefaultRoles(ctx context.Context) error {
    // 1. Ensure permissions exist
    perms := []string{
        "user:create", "user:read", "user:update", "user:delete",
        "role:create", "role:read", "role:update", "role:delete",
        "room:create", "room:read", "room:update", "room:delete",
        "recording:read", "recording:delete",
        "group:manage",
    }
    
    for _, k := range perms {
        _, err := s.repo.FindPermissionByKey(ctx, k)
        if err != nil {
             s.repo.CreatePermission(ctx, &role.Permission{Key: k, Description: "Auto-generated"})
        }
    }
    
    // 2. Ensure roles exist
    adminRole, err := s.repo.FindByName(ctx, "admin")
    if err != nil || adminRole == nil {
        adminRole, _ = s.CreateRole(ctx, "admin", "Administrator")
        // Assign all perms to admin
        allPerms, _ := s.repo.ListPermissions(ctx)
        for _, p := range allPerms {
            s.repo.AssignPermission(ctx, adminRole.ID, p.ID)
        }
    }
    
    userRole, err := s.repo.FindByName(ctx, "user")
    if err != nil || userRole == nil {
        s.CreateRole(ctx, "user", "Standard User")
    }
    
    return nil
}
