package service

import (
	"errors"

	"video-conference-be/internal/domain/user"
	"video-conference-be/internal/domain/role"
	"video-conference-be/pkg/utility"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	ListUsers() ([]user.User, error)
	CreateUser(username, password string, roleID uint) (*user.User, error)
	UpdateUserRole(id uint, roleID uint) (*user.User, error)
	DeleteUser(id uint) error
}

type userService struct {
    // We might need RoleService here if we want to look up roles by name easily
    // Or just use DB directly? Better to use RoleService or Repo.
    // However, clean architecture says Service shouldn't depend on another Service ideally, but pragmatic approach allows it.
    // Or better, RoleRepo. 
    // Let's inject generic DB access via utility.DB or similar if that's the pattern used.
    // The previous code used utility.DB directly. 
}

func NewUserService() UserService {
	return &userService{}
}

func (s *userService) ListUsers() ([]user.User, error) {
	var users []user.User
	if err := utility.DB.
		Preload("Role"). /* Preload role data */
		Select("id", "username", "role_id", "created_at", "updated_at").
		Order("id ASC").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *userService) CreateUser(username, password string, roleID uint) (*user.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password are required")
	}

    // Validate Role
    var r role.Role
    if roleID == 0 {
         // Try to find default "user" role if roleID is 0? Or just error?
         // Let's force a valid ID or find "user"
         if err := utility.DB.Where("name = ?", "user").First(&r).Error; err != nil {
             return nil, errors.New("default role 'user' not found, please provide a valid role_id")
         }
         roleID = r.ID
    } else {
        if err := utility.DB.First(&r, roleID).Error; err != nil {
            return nil, errors.New("role not found")
        }
    }

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &user.User{
		Username:     username,
		PasswordHash: string(hashed),
		RoleID:       roleID,
	}

	if err := utility.DB.Create(u).Error; err != nil {
		return nil, err
	}

	u.PasswordHash = "" // safety
    u.Role = &r
	return u, nil
}

func (s *userService) UpdateUserRole(id uint, roleID uint) (*user.User, error) {
	if roleID == 0 {
		return nil, errors.New("role_id is required")
	}

	var u user.User
	if err := utility.DB.First(&u, id).Error; err != nil {
		return nil, err
	}
    
    var r role.Role
    if err := utility.DB.First(&r, roleID).Error; err != nil {
        return nil, errors.New("role not found")
    }

    u.RoleID = r.ID
	if err := utility.DB.Save(&u).Error; err != nil {
		return nil, err
	}
    // reload with role
    utility.DB.Preload("Role").First(&u, u.ID)

	u.PasswordHash = ""
	return &u, nil
}

func (s *userService) DeleteUser(id uint) error {
	// Manually delete from group_members to avoid foreign key constraint errors
	if err := utility.DB.Table("group_members").Where("user_id = ?", id).Delete(nil).Error; err != nil {
		return err
	}

	if err := utility.DB.Delete(&user.User{}, id).Error; err != nil {
		return err
	}
	return nil
}
