package model

import "time"

type RegisterRequest struct {
	Username     string  `json:"username"`
	Email        *string `json:"email"`
	Password     string  `json:"password"`
	DisplayName  *string `json:"display_name"`
	AvatarURL    *string `json:"avatar_url"`
	AuthProvider string  `json:"auth_provider"`
	GoogleSub    *string `json:"google_sub"`
	TelegramChat *string `json:"telegram_chat"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type AssignRoleRequest struct {
	RoleID uint `json:"role_id"`
}

type UpdateUserRequest struct {
	DisplayName  *string `json:"display_name"`
	AvatarURL    *string `json:"avatar_url"`
	Email        *string `json:"email"`
	TelegramChat *string `json:"telegram_chat"`
	Role         *string `json:"role"`
	IsDeleted    *bool   `json:"is_deleted"`
}

type LoginResponse struct {
	Token     string      `json:"token"`
	User      interface{} `json:"user"`
	ExpiresAt time.Time   `json:"expires_at"`
}
