package service

import (
	"errors"
	"strings"

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
	username := strings.ToLower(strings.TrimSpace(req.Username))
	if username == "" || strings.TrimSpace(req.Password) == "" {
		return nil, errors.New("username and password are required")
	}
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}
	if _, err := s.repo.FindUserByUsername(username); err == nil {
		return nil, errors.New("username already registered")
	}

	normalizedEmail := normalizedOptional(req.Email, true)
	if normalizedEmail != nil {
		if _, err := s.repo.FindUserByEmail(*normalizedEmail); err == nil {
			return nil, errors.New("email already registered")
		}
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	provider := strings.ToLower(strings.TrimSpace(req.AuthProvider))
	if provider == "" {
		provider = "local"
	}

	googleSub := normalizedOptional(req.GoogleSub, false)
	if provider == "google" && googleSub == nil {
		return nil, errors.New("google_sub is required for google auth provider")
	}

	user := &database.User{
		Username:     username,
		Email:        normalizedEmail,
		Password:     string(hashed),
		DisplayName:  normalizedOptional(req.DisplayName, false),
		AvatarURL:    normalizedOptional(req.AvatarURL, false),
		AuthProvider: provider,
		GoogleSub:    googleSub,
		TelegramChat: normalizedOptional(req.TelegramChat, false),
		Role:         "user",
		IsDeleted:    false,
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
	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" || strings.TrimSpace(req.Password) == "" {
		return nil, errors.New("identifier and password are required")
	}

	user, err := s.repo.FindUserByIdentifier(identifier)
	if err != nil {
		return nil, errors.New("invalid identifier or password")
	}

	if user.IsDeleted {
		return nil, errors.New("user account is deleted")
	}
	if user.AuthProvider != "local" {
		return nil, errors.New("password login is not enabled for this account")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid identifier or password")
	}

	roles, err := s.repo.UserRoles(user.ID)
	if err != nil {
		return nil, err
	}
	permissions, err := s.repo.UserPermissions(user.ID)
	if err != nil {
		return nil, err
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	token, expiresAt, err := s.jwt.Generate(user.ID, email, domain, permissions, roles)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return &model.LoginResponse{Token: token, User: user, ExpiresAt: expiresAt}, nil
}

func (s *Service) ChangePassword(userID string, oldPassword, newPassword string) error {
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
	return s.repo.UpdateUser(user)
}

func (s *Service) Me(userID string) (map[string]interface{}, error) {
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

func (s *Service) GetUser(userID string) (map[string]interface{}, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	user.Password = ""
	roles, _ := s.repo.UserRoles(userID)
	return map[string]interface{}{"user": user, "roles": roles}, nil
}

func (s *Service) UpdateUser(userID string, req model.UpdateUserRequest) error {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if req.DisplayName != nil {
		user.DisplayName = normalizedOptional(req.DisplayName, false)
	}
	if req.AvatarURL != nil {
		user.AvatarURL = normalizedOptional(req.AvatarURL, false)
	}
	if req.Email != nil {
		normalized := normalizedOptional(req.Email, true)
		if normalized != nil {
			existing, err := s.repo.FindUserByEmail(*normalized)
			if err == nil && existing.ID != user.ID {
				return errors.New("email already registered")
			}
		}
		user.Email = normalized
	}
	if req.TelegramChat != nil {
		user.TelegramChat = normalizedOptional(req.TelegramChat, false)
	}
	if req.Role != nil {
		role := strings.TrimSpace(strings.ToLower(*req.Role))
		if role != "" {
			user.Role = role
		}
	}
	if req.IsDeleted != nil {
		user.IsDeleted = *req.IsDeleted
	}

	return s.repo.UpdateUser(user)
}

func (s *Service) DeleteUser(userID string) error {
	return s.repo.DeleteUser(userID)
}

func (s *Service) AssignRole(userID string, roleID uint) error {
	if _, err := s.repo.FindRoleByID(roleID); err != nil {
		return errors.New("role not found")
	}
	return s.repo.AssignRole(userID, roleID)
}

func (s *Service) RemoveRole(userID string, roleID uint) error {
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

func (s *Service) HasPermission(userID string, permission string) (bool, error) {
	return s.repo.HasPermission(userID, permission)
}

func normalizedOptional(v *string, toLower bool) *string {
	if v == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil
	}
	if toLower {
		trimmed = strings.ToLower(trimmed)
	}
	return &trimmed
}
