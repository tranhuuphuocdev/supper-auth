package service

import (
	"errors"
	"strings"
	"time"

	"auth-service/internal/core/database"
	"auth-service/internal/core/jwtx"
	"auth-service/internal/modules/auth/model"
	"auth-service/internal/modules/auth/repository"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *repository.Repository
	jwt  *jwtx.Service
}

func New(repo *repository.Repository, jwtService *jwtx.Service) *Service {
	return &Service{repo: repo, jwt: jwtService}
}

func (s *Service) Register(req model.RegisterRequest) (*database.User, error) {
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, errors.New("email and password are required")
	}
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	if _, err := s.repo.FindUserByEmail(req.Email); err == nil {
		return nil, errors.New("email already registered")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &database.User{
		Email:     strings.ToLower(strings.TrimSpace(req.Email)),
		Password:  string(hashed),
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		IsActive:  true,
	}
	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	role, err := s.repo.FindRoleByName("user")
	if err == nil {
		_ = s.repo.AssignRole(user.ID, role.ID)
	}

	user.Password = ""
	return user, nil
}

func (s *Service) Login(domain string, req model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.repo.FindUserByEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	roles, err := s.repo.UserRoles(user.ID)
	if err != nil {
		return nil, err
	}
	permissions, err := s.repo.UserPermissions(user.ID)
	if err != nil {
		return nil, err
	}

	token, expiresAt, err := s.jwt.Generate(user.ID, user.Email, domain, permissions, roles)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return &model.LoginResponse{Token: token, User: user, ExpiresAt: expiresAt}, nil
}

func (s *Service) ChangePassword(userID uint, oldPassword, newPassword string) error {
	if len(strings.TrimSpace(newPassword)) < 6 {
		return errors.New("new password must be at least 6 characters")
	}

	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("invalid password")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashed)
	user.UpdatedAt = time.Now()
	return s.repo.UpdateUser(user)
}

func (s *Service) Me(userID uint) (map[string]interface{}, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	user.Password = ""

	roles, _ := s.repo.UserRoles(userID)
	permissions, _ := s.repo.UserPermissions(userID)

	return map[string]interface{}{
		"user":        user,
		"roles":       roles,
		"permissions": permissions,
	}, nil
}

func (s *Service) ListUsers() ([]database.User, error) {
	users, err := s.repo.ListUsers()
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i].Password = ""
	}
	return users, nil
}

func (s *Service) GetUser(userID uint) (map[string]interface{}, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	user.Password = ""
	roles, _ := s.repo.UserRoles(userID)
	return map[string]interface{}{"user": user, "roles": roles}, nil
}

func (s *Service) UpdateUser(userID uint, req model.UpdateUserRequest) error {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}
	user.FirstName = strings.TrimSpace(req.FirstName)
	user.LastName = strings.TrimSpace(req.LastName)
	user.IsActive = req.IsActive
	return s.repo.UpdateUser(user)
}

func (s *Service) DeleteUser(userID uint) error {
	return s.repo.DeleteUser(userID)
}

func (s *Service) AssignRole(userID, roleID uint) error {
	if _, err := s.repo.FindRoleByID(roleID); err != nil {
		return errors.New("role not found")
	}
	return s.repo.AssignRole(userID, roleID)
}

func (s *Service) RemoveRole(userID, roleID uint) error {
	return s.repo.RemoveRole(userID, roleID)
}

func (s *Service) ListRoles() ([]database.Role, error) {
	return s.repo.ListRoles()
}

func (s *Service) CreateRole(name, description string) (*database.Role, error) {
	role := &database.Role{Name: strings.TrimSpace(name), Description: strings.TrimSpace(description)}
	if role.Name == "" {
		return nil, errors.New("role name is required")
	}
	if err := s.repo.CreateRole(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *Service) ListPermissions() ([]database.Permission, error) {
	return s.repo.ListPermissions()
}

func (s *Service) HasPermission(userID uint, permission string) (bool, error) {
	return s.repo.HasPermission(userID, permission)
}
