package service

import (
	"errors"

	"video-conference-be/internal/domain/user"
	"video-conference-be/pkg/utility"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	ListUsers() ([]user.User, error)
	CreateUser(username, password, role string) (*user.User, error)
	UpdateUserRole(id uint, role string) (*user.User, error)
	DeleteUser(id uint) error
}

type userService struct{}

func NewUserService() UserService {
	return &userService{}
}

func (s *userService) ListUsers() ([]user.User, error) {
	var users []user.User
	if err := utility.DB.
		Select("id", "username", "role", "created_at", "updated_at").
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

	var role user.Role
	switch user.Role(roleStr) {
	case user.RoleAdmin:
		role = user.RoleAdmin
	default:
		role = user.RoleUser
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &user.User{
		Username:     username,
		PasswordHash: string(hashed),
		Role:         role,
	}

	if err := utility.DB.Create(u).Error; err != nil {
		return nil, err
	}

	u.PasswordHash = "" // safety
	return u, nil
}

func (s *userService) UpdateUserRole(id uint, roleStr string) (*user.User, error) {
	if roleStr == "" {
		return nil, errors.New("role is required")
	}

	var u user.User
	if err := utility.DB.First(&u, id).Error; err != nil {
		return nil, err
	}

	switch user.Role(roleStr) {
	case user.RoleAdmin:
		u.Role = user.RoleAdmin
	default:
		u.Role = user.RoleUser
	}

	if err := utility.DB.Save(&u).Error; err != nil {
		return nil, err
	}

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
