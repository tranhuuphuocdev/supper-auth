package database

import (
	"fmt"
	"sync"

	"auth-service/internal/core/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Manager struct {
	cfg *config.Config

	mu            sync.RWMutex
	dbByDomain    map[string]*gorm.DB
	domainBootstr map[string]bool
}

func NewManager(cfg *config.Config) (*Manager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	return &Manager{
		cfg:           cfg,
		dbByDomain:    make(map[string]*gorm.DB),
		domainBootstr: make(map[string]bool),
	}, nil
}

func (m *Manager) DBForDomain(domain string) (*gorm.DB, string, error) {
	resolvedDomain, dc := m.cfg.ResolveDomain(domain)

	m.mu.RLock()
	existing, ok := m.dbByDomain[resolvedDomain]
	m.mu.RUnlock()
	if ok {
		return existing, resolvedDomain, nil
	}

	db, err := gorm.Open(postgres.Open(buildDSN(m.cfg, dc)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, "", fmt.Errorf("connect domain %s: %w", resolvedDomain, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, "", err
	}

	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetMaxIdleConns(10)

	m.mu.Lock()
	if current, exists := m.dbByDomain[resolvedDomain]; exists {
		m.mu.Unlock()
		_ = sqlDB.Close()
		return current, resolvedDomain, nil
	}
	m.dbByDomain[resolvedDomain] = db
	m.mu.Unlock()

	return db, resolvedDomain, nil
}

func (m *Manager) EnsureDomainReady(domain string) error {
	db, resolved, err := m.DBForDomain(domain)
	if err != nil {
		return err
	}

	m.mu.RLock()
	alreadyBooted := m.domainBootstr[resolved]
	m.mu.RUnlock()
	if alreadyBooted {
		return nil
	}

	if err := migrate(db); err != nil {
		return fmt.Errorf("migrate domain %s: %w", resolved, err)
	}
	if err := seedDefaultData(db); err != nil {
		return fmt.Errorf("seed domain %s: %w", resolved, err)
	}

	m.mu.Lock()
	m.domainBootstr[resolved] = true
	m.mu.Unlock()

	return nil
}

func (m *Manager) CloseAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for domain, db := range m.dbByDomain {
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("close domain %s: %w", domain, err)
		}
		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("close domain %s: %w", domain, err)
		}
	}
	return nil
}

func buildDSN(cfg *config.Config, dc config.DomainConfig) string {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		dc.DBName,
		cfg.DBSSLMode,
	)

	if dc.Schema != "" {
		dsn += fmt.Sprintf(" search_path=%s", dc.Schema)
	}

	return dsn
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Role{}, &Permission{}, &UserRole{}, &RolePermission{})
}

func seedDefaultData(db *gorm.DB) error {
	roles := []Role{
		{Name: "admin", Description: "Administrator with full access"},
		{Name: "moderator", Description: "Moderator with limited management access"},
		{Name: "user", Description: "Regular user"},
	}

	permissions := []Permission{
		{Name: "user.create", Description: "Can create users"},
		{Name: "user.read", Description: "Can read user data"},
		{Name: "user.update", Description: "Can update users"},
		{Name: "user.delete", Description: "Can delete users"},
		{Name: "role.manage", Description: "Can manage roles"},
		{Name: "permission.manage", Description: "Can manage permissions"},
	}

	for i := range roles {
		if err := db.Where(Role{Name: roles[i].Name}).FirstOrCreate(&roles[i]).Error; err != nil {
			return err
		}
	}
	for i := range permissions {
		if err := db.Where(Permission{Name: permissions[i].Name}).FirstOrCreate(&permissions[i]).Error; err != nil {
			return err
		}
	}

	byRole := map[string][]string{
		"admin":     {"user.create", "user.read", "user.update", "user.delete", "role.manage", "permission.manage"},
		"moderator": {"user.read", "user.update"},
		"user":      {"user.read"},
	}

	roleByName := map[string]Role{}
	for _, role := range roles {
		roleByName[role.Name] = role
	}
	permByName := map[string]Permission{}
	for _, p := range permissions {
		permByName[p.Name] = p
	}

	for roleName, permissionNames := range byRole {
		role := roleByName[roleName]
		for _, pn := range permissionNames {
			perm := permByName[pn]
			link := RolePermission{RoleID: role.ID, PermissionID: perm.ID}
			if err := db.Where(link).FirstOrCreate(&link).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
