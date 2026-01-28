package service

import (
    "context"
    dUser "video-conference-be/internal/domain/user"
)

func (s *authService) SyncUserFromSSO(ctx context.Context, username, email string) (*dUser.User, error) {
	// 1. Try to find user by username
	u, err := s.userRepo.FindByUsername(ctx, username)
	if err == nil && u != nil {
		return u, nil
	}

	// 2. If not found, create new user
	// Check for "user" role
	// defaultRole, err := s.roleService.ListRoles(ctx) // optimization: add FindByName to RoleService
    // Just list roles directly

	// Or better, assume FindByName exists if I add it.
	// For now, let's just get the first role or "user" role if I implement FindByName in Service
    // I implemented FindByName in Repo but not Service. I should add FindByName to Service or just list all.
    // Let's use ListRoles and find "user".
    
    // Quick Fix: I'll assume RoleService has FindByName or I iterate.
    // I will iterate for now to be safe.
    var roleID uint
    roles, _ := s.roleService.ListRoles(ctx)
    for _, r := range roles {
        if r.Name == "user" {
            roleID = r.ID
            break
        }
    }
    
    // Fallback if no "user" role, perform init? Or pick first?
    if roleID == 0 && len(roles) > 0 {
        roleID = roles[0].ID
    }
    
    // Create User
    newUser := &dUser.User{
        Username: username,
        RoleID:   roleID,
    }
    
    // Note: PasswordHash is empty for SSO users
    if err := s.userRepo.Create(ctx, newUser); err != nil {
        return nil, err
    }
    
    // Reload user to get Role populated
    return s.userRepo.FindByID(ctx, newUser.ID)
}
