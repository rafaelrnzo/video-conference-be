package service

import (
	"errors"
	"video-conference-be/internal/app/repository"
	"video-conference-be/internal/domain/role"

)

type RoleService interface {
	CreateRole(name string) (*role.Role, error)
	ListRoles() ([]string, error)
}

type roleService struct {
	repo repository.RoleRepository
}

func NewRoleService(repo repository.RoleRepository) RoleService {
	return &roleService{repo: repo}
}

func (s *roleService) CreateRole(name string) (*role.Role, error) {
	if name == "" {
		return nil, errors.New("role name required")
	}

	existing, _ := s.repo.FindByName(name)
	if existing != nil && existing.ID != 0 {
		return existing, nil // idempotency
	}

	newRole := &role.Role{
		Name: name,
	}

	if err := s.repo.Create(newRole); err != nil {
		return nil, err
	}

	return newRole, nil
}

func (s *roleService) ListRoles() ([]string, error) {
	// fetching from DB
	roles, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}

	unique := make(map[string]bool)
	for _, r := range roles {
		unique[r.Name] = true
	}

	var result []string
	for r := range unique {
		result = append(result, r)
	}

	return result, nil
}
