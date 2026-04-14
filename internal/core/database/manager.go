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
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto").Error; err != nil {
		return err
	}
	if err := migrateLegacyUsers(db); err != nil {
		return err
	}
	if err := migrateLegacyEpochColumns(db); err != nil {
		return err
	}
	if err := ensureLegacyNamedConstraints(db); err != nil {
		return err
	}
	return db.AutoMigrate(&User{}, &Role{}, &Permission{}, &UserRole{}, &RolePermission{})
}

func ensureLegacyNamedConstraints(db *gorm.DB) error {
	queries := []string{
		`
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = current_schema() AND table_name = 'users' AND column_name = 'username'
			) AND NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'uni_users_username'
			) THEN
				BEGIN
					ALTER TABLE users ADD CONSTRAINT uni_users_username UNIQUE (username);
				EXCEPTION
					WHEN unique_violation THEN
						WITH dup AS (
							SELECT ctid, username, ROW_NUMBER() OVER (PARTITION BY username ORDER BY ctid) AS rn
							FROM users
						)
						UPDATE users u
						SET username = CONCAT(dup.username, '_', dup.rn)
						FROM dup
						WHERE u.ctid = dup.ctid AND dup.rn > 1;

						ALTER TABLE users ADD CONSTRAINT uni_users_username UNIQUE (username);
					WHEN duplicate_object THEN
						NULL;
				END;
			END IF;
		END
		$$;
		`,
		`
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = current_schema() AND table_name = 'users' AND column_name = 'google_sub'
			) AND NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'uni_users_google_sub'
			) THEN
				BEGIN
					ALTER TABLE users ADD CONSTRAINT uni_users_google_sub UNIQUE (google_sub);
				EXCEPTION
					WHEN unique_violation THEN
						WITH dup AS (
							SELECT ctid, ROW_NUMBER() OVER (PARTITION BY google_sub ORDER BY ctid) AS rn
							FROM users
							WHERE google_sub IS NOT NULL
						)
						UPDATE users u
						SET google_sub = NULL
						FROM dup
						WHERE u.ctid = dup.ctid AND dup.rn > 1;

						ALTER TABLE users ADD CONSTRAINT uni_users_google_sub UNIQUE (google_sub);
					WHEN duplicate_object THEN
						NULL;
				END;
			END IF;
		END
		$$;
		`,
		`
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = current_schema() AND table_name = 'roles' AND column_name = 'name'
			) AND NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'uni_roles_name'
			) THEN
				BEGIN
					ALTER TABLE roles ADD CONSTRAINT uni_roles_name UNIQUE (name);
				EXCEPTION
					WHEN unique_violation THEN
						WITH dup AS (
							SELECT ctid, name, ROW_NUMBER() OVER (PARTITION BY name ORDER BY ctid) AS rn
							FROM roles
						)
						UPDATE roles r
						SET name = CONCAT(dup.name, '_', dup.rn)
						FROM dup
						WHERE r.ctid = dup.ctid AND dup.rn > 1;

						ALTER TABLE roles ADD CONSTRAINT uni_roles_name UNIQUE (name);
					WHEN duplicate_object THEN
						NULL;
				END;
			END IF;
		END
		$$;
		`,
		`
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = current_schema() AND table_name = 'permissions' AND column_name = 'name'
			) AND NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'uni_permissions_name'
			) THEN
				BEGIN
					ALTER TABLE permissions ADD CONSTRAINT uni_permissions_name UNIQUE (name);
				EXCEPTION
					WHEN unique_violation THEN
						WITH dup AS (
							SELECT ctid, name, ROW_NUMBER() OVER (PARTITION BY name ORDER BY ctid) AS rn
							FROM permissions
						)
						UPDATE permissions p
						SET name = CONCAT(dup.name, '_', dup.rn)
						FROM dup
						WHERE p.ctid = dup.ctid AND dup.rn > 1;

						ALTER TABLE permissions ADD CONSTRAINT uni_permissions_name UNIQUE (name);
					WHEN duplicate_object THEN
						NULL;
				END;
			END IF;
		END
		$$;
		`,
	}

	for _, q := range queries {
		if err := db.Exec(q).Error; err != nil {
			return err
		}
	}

	return nil
}

func migrateLegacyEpochColumns(db *gorm.DB) error {
	targets := [][2]string{
		{"users", "created_at"},
		{"users", "updated_at"},
		{"roles", "created_at"},
		{"roles", "updated_at"},
		{"permissions", "created_at"},
		{"permissions", "updated_at"},
	}

	for _, t := range targets {
		if err := convertTimestampColumnToEpochMillis(db, t[0], t[1]); err != nil {
			return err
		}
	}

	return nil
}

func convertTimestampColumnToEpochMillis(db *gorm.DB, tableName string, columnName string) error {
	return db.Exec(fmt.Sprintf(`
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = current_schema()
				  AND table_name = '%s'
				  AND column_name = '%s'
				  AND data_type IN ('timestamp without time zone', 'timestamp with time zone')
			) THEN
				EXECUTE 'ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT';
				EXECUTE 'ALTER TABLE %s ALTER COLUMN %s TYPE bigint USING (EXTRACT(EPOCH FROM %s) * 1000)::bigint';
			END IF;
		END
		$$;
	`, tableName, columnName, tableName, columnName, tableName, columnName, columnName)).Error
}

func migrateLegacyUsers(db *gorm.DB) error {
	if !db.Migrator().HasTable("users") {
		return nil
	}

	if !db.Migrator().HasColumn(&User{}, "u_id") {
		if err := db.Exec("ALTER TABLE users ADD COLUMN u_id uuid").Error; err != nil {
			return err
		}
	}
	if err := db.Exec("ALTER TABLE users ALTER COLUMN u_id SET DEFAULT gen_random_uuid()").Error; err != nil {
		return err
	}
	if err := db.Exec("UPDATE users SET u_id = gen_random_uuid() WHERE u_id IS NULL").Error; err != nil {
		return err
	}

	if !db.Migrator().HasColumn(&User{}, "username") {
		if err := db.Exec("ALTER TABLE users ADD COLUMN username text").Error; err != nil {
			return err
		}
	}

	if err := db.Exec(`
		UPDATE users
		SET username = NULLIF(
			LOWER(
				REGEXP_REPLACE(
					SPLIT_PART(COALESCE(email, ''), '@', 1),
					'[^a-zA-Z0-9_]+',
					'',
					'g'
				)
			),
			''
		)
		WHERE username IS NULL OR BTRIM(username) = ''
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		UPDATE users
		SET username = CONCAT('user_', SUBSTRING(REPLACE(u_id::text, '-', ''), 1, 12))
		WHERE username IS NULL OR BTRIM(username) = ''
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		WITH dup AS (
			SELECT ctid, username, ROW_NUMBER() OVER (PARTITION BY username ORDER BY ctid) AS rn
			FROM users
		)
		UPDATE users u
		SET username = CONCAT(dup.username, '_', dup.rn)
		FROM dup
		WHERE u.ctid = dup.ctid
		  AND dup.rn > 1
	`).Error; err != nil {
		return err
	}

	if err := ensureLegacyEmailConstraint(db); err != nil {
		return err
	}

	return nil
}

func ensureLegacyEmailConstraint(db *gorm.DB) error {
	return db.Exec(`
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = current_schema()
				  AND table_name = 'users'
				  AND column_name = 'email'
			) THEN
				IF NOT EXISTS (
					SELECT 1
					FROM pg_constraint
					WHERE conname = 'uni_users_email'
				) THEN
					BEGIN
						ALTER TABLE users ADD CONSTRAINT uni_users_email UNIQUE (email);
					EXCEPTION
						WHEN unique_violation THEN
							WITH d AS (
								SELECT ctid, ROW_NUMBER() OVER (PARTITION BY LOWER(email) ORDER BY ctid) AS rn
								FROM users
								WHERE email IS NOT NULL
							)
							UPDATE users u
							SET email = NULL
							FROM d
							WHERE u.ctid = d.ctid
							  AND d.rn > 1;

							ALTER TABLE users ADD CONSTRAINT uni_users_email UNIQUE (email);
						WHEN duplicate_object THEN
							NULL;
					END;
				END IF;
			END IF;
		END
		$$;
	`).Error
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
