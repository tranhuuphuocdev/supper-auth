package database

type User struct {
	ID           string  `gorm:"type:uuid;default:gen_random_uuid();primaryKey;column:u_id" json:"id"`
	DisplayName  *string `gorm:"column:dn;size:255" json:"display_name,omitempty"`
	AvatarURL    *string `gorm:"column:avatar_url;type:text" json:"avatar_url,omitempty"`
	Username     string  `gorm:"column:username;size:100;uniqueIndex;not null" json:"username"`
	Email        *string `gorm:"column:email;size:255;uniqueIndex" json:"email,omitempty"`
	AuthProvider string  `gorm:"column:auth_provider;size:50;not null;default:local" json:"auth_provider"`
	GoogleSub    *string `gorm:"column:google_sub;size:255;uniqueIndex" json:"google_sub,omitempty"`
	TelegramChat *string `gorm:"column:tele_chat_id;size:255" json:"telegram_chat,omitempty"`
	Password     string  `gorm:"column:password;size:255;not null" json:"-"`
	Role         string  `gorm:"column:role;size:50;not null;default:user" json:"role"`
	CreatedAt    int64   `gorm:"column:created_at;autoCreateTime:milli" json:"created_at"`
	UpdatedAt    int64   `gorm:"column:updated_at;autoUpdateTime:milli" json:"updated_at"`
	IsDeleted    bool    `gorm:"column:is_deleted;not null;default:false" json:"is_deleted"`

	Roles []Role `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

func (User) TableName() string {
	return "users"
}

type Role struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`

	Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}

type Permission struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:100;uniqueIndex;not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
	CreatedAt   int64  `gorm:"autoCreateTime:milli" json:"created_at"`
	UpdatedAt   int64  `gorm:"autoUpdateTime:milli" json:"updated_at"`
}

type UserRole struct {
	UserID string `gorm:"type:uuid;primaryKey;column:user_id"`
	RoleID uint   `gorm:"primaryKey;column:role_id"`
}

type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}
