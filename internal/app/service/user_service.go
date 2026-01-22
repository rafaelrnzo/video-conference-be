package service

import (
	"errors"

	"video-conference-be/internal/domain/user"
	"video-conference-be/internal/pkg/rbac"
	"video-conference-be/pkg/utility"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	ListUsers() ([]user.User, error)
	CreateUser(username, password, role string) (*user.User, error)
	UpdateUserRole(id uint, role string) (*user.User, error)
	DeleteUser(id uint) error
}


type userService struct {
	roleSvc RoleService
}

func NewUserService(roleSvc RoleService) UserService {
	return &userService{roleSvc: roleSvc}
}

func (s *userService) ListUsers() ([]user.User, error) {
	var users []user.User
	if err := utility.DB.
		Preload("Role").
		Order("id ASC").
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *userService) CreateUser(username, password, roleStr string) (*user.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password are required")
	}

	// Dynamic role: just accept the string.
	roleVal := roleStr
	if roleVal == "" {
		roleVal = "user"
	}

	// Ensure role exists in DB and get it
	roleEntity, err := s.roleSvc.CreateRole(roleVal)
	if err != nil {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &user.User{
		Username:     username,
		PasswordHash: string(hashed),
		RoleID:       roleEntity.ID,
		Role:         *roleEntity,
	}

	if err := utility.DB.Create(u).Error; err != nil {
		return nil, err
	}

	// Sync with Casbin
	_, _ = rbac.Enforcer.AddGroupingPolicy(u.Username, u.Role.Name)

	u.PasswordHash = "" // safety
	return u, nil
}

func (s *userService) UpdateUserRole(id uint, roleStr string) (*user.User, error) {
	if roleStr == "" {
		return nil, errors.New("role is required")
	}

	// Ensure role exists in DB
	roleEntity, err := s.roleSvc.CreateRole(roleStr)
	if err != nil {
		return nil, err
	}

	var u user.User
	// Preload role to get old role name if needed
	if err := utility.DB.Preload("Role").First(&u, id).Error; err != nil {
		return nil, err
	}

	oldRole := u.Role.Name
	newRole := roleStr

	u.RoleID = roleEntity.ID
	u.Role = *roleEntity

	if err := utility.DB.Save(&u).Error; err != nil {
		return nil, err
	}

	// Sync with Casbin
	if oldRole != "" {
		_, _ = rbac.Enforcer.RemoveGroupingPolicy(u.Username, oldRole)
	}
	_, _ = rbac.Enforcer.AddGroupingPolicy(u.Username, newRole)

	u.PasswordHash = ""
	return &u, nil
}

func (s *userService) DeleteUser(id uint) error {
	var u user.User
	if err := utility.DB.First(&u, id).Error; err == nil {
		// Remove from Casbin
		_, _ = rbac.Enforcer.RemoveFilteredGroupingPolicy(0, u.Username)
	}

	// Manually delete from group_members to avoid foreign key constraint errors
	if err := utility.DB.Table("group_members").Where("user_id = ?", id).Delete(nil).Error; err != nil {
		return err
	}

	if err := utility.DB.Delete(&user.User{}, id).Error; err != nil {
		return err
	}
	return nil
}
