package role

import "context"

type Repository interface {
	Create(ctx context.Context, role *Role) error
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*Role, error)
	FindByName(ctx context.Context, name string) (*Role, error)
	List(ctx context.Context) ([]Role, error)
	
	// Permission management
	AssignPermission(ctx context.Context, roleID, permID uint) error
	RevokePermission(ctx context.Context, roleID, permID uint) error
	ListPermissions(ctx context.Context) ([]Permission, error)
    CreatePermission(ctx context.Context, perm *Permission) error
    FindPermissionByKey(ctx context.Context, key string) (*Permission, error)
}
