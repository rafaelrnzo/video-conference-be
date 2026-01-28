package service

import (
	"context"
	"errors"
	"time"

	dUser "video-conference-be/internal/domain/user"
	"video-conference-be/pkg/utility"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, username, password string) (*dUser.User, error)
	Login(ctx context.Context, username, password string) (string, *dUser.User, error)
	SyncUserFromSSO(ctx context.Context, username, email string) (*dUser.User, error)
}

type authService struct {
	userRepo   dUser.Repository
	roleService RoleService
}

func NewAuthService(userRepo dUser.Repository, roleService RoleService) AuthService {
	return &authService{userRepo: userRepo, roleService: roleService}
}

func (s *authService) Register(ctx context.Context, username, password string) (*dUser.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

    // Find default role
    var roleID uint
    roles, _ := s.roleService.ListRoles(ctx)
    for _, r := range roles {
        if r.Name == "user" {
            roleID = r.ID
            break
        }
    }
    if roleID == 0 {
        // Create if missing? Or error?
        // Let's create default roles if missing
        s.roleService.InitDefaultRoles(ctx)
        roles, _ = s.roleService.ListRoles(ctx)
         for _, r := range roles {
            if r.Name == "user" {
                roleID = r.ID
                break
            }
        }
    }

	u := &dUser.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		RoleID:       roleID,
	}
	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *authService) Login(ctx context.Context, username, password string) (string, *dUser.User, error) {
	u, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}
	roleName := "user"
	if u.Role != nil {
		roleName = u.Role.Name
	}
	token, err := utility.GenerateJWT(u.Username, roleName, 24*time.Hour)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}
