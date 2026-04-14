package repository

import (
	"errors"
	"strings"

	"auth-service/internal/core/database"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(user *database.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) FindUserByEmail(email string) (*database.User, error) {
	var user database.User
	normalized := strings.ToLower(strings.TrimSpace(email))
	if err := r.db.Where("LOWER(email) = ? AND is_deleted = ?", normalized, false).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindUserByUsername(username string) (*database.User, error) {
	var user database.User
	normalized := strings.ToLower(strings.TrimSpace(username))
	if err := r.db.Where("LOWER(username) = ? AND is_deleted = ?", normalized, false).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindUserByIdentifier(identifier string) (*database.User, error) {
	var user database.User
	normalized := strings.ToLower(strings.TrimSpace(identifier))
	if err := r.db.
		Where("(LOWER(username) = ? OR LOWER(email) = ?) AND is_deleted = ?", normalized, normalized, false).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindUserByID(userID string) (*database.User, error) {
	var user database.User
	if err := r.db.Where("u_id = ? AND is_deleted = ?", userID, false).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUser(user *database.User) error {
	return r.db.Save(user).Error
}

func (r *Repository) DeleteUser(userID string) error {
	return r.db.Model(&database.User{}).Where("u_id = ?", userID).Update("is_deleted", true).Error
}

func (r *Repository) ListUsers() ([]database.User, error) {
	var users []database.User
	if err := r.db.Where("is_deleted = ?", false).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *Repository) ListRoles() ([]database.Role, error) {
	var roles []database.Role
	if err := r.db.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *Repository) ListPermissions() ([]database.Permission, error) {
	var permissions []database.Permission
	if err := r.db.Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *Repository) CreateRole(role *database.Role) error {
	return r.db.Create(role).Error
}

func (r *Repository) FindRoleByName(name string) (*database.Role, error) {
	var role database.Role
	if err := r.db.Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repository) FindRoleByID(roleID uint) (*database.Role, error) {
	var role database.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *Repository) AssignRole(userID string, roleID uint) error {
	link := database.UserRole{UserID: userID, RoleID: roleID}
	return r.db.Where(link).FirstOrCreate(&link).Error
}

func (r *Repository) RemoveRole(userID string, roleID uint) error {
	return r.db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&database.UserRole{}).Error
}

func (r *Repository) UserRoles(userID string) ([]string, error) {
	var roles []string
	err := r.db.Table("roles").
		Select("roles.name").
		Joins("JOIN user_roles ur ON ur.role_id = roles.id").
		Where("ur.user_id = ?", userID).
		Scan(&roles).Error
	return roles, err
}

func (r *Repository) UserPermissions(userID string) ([]string, error) {
	var permissions []string
	err := r.db.Table("permissions").
		Distinct("permissions.name").
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Joins("JOIN user_roles ur ON ur.role_id = rp.role_id").
		Where("ur.user_id = ?", userID).
		Scan(&permissions).Error
	return permissions, err
}

func (r *Repository) HasPermission(userID string, permission string) (bool, error) {
	var count int64
	err := r.db.Table("permissions").
		Joins("JOIN role_permissions rp ON rp.permission_id = permissions.id").
		Joins("JOIN user_roles ur ON ur.role_id = rp.role_id").
		Where("ur.user_id = ? AND permissions.name = ?", userID, permission).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
