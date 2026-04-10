package jwtx

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	secret    []byte
	expiresIn time.Duration
}

type Claims struct {
	UserID      uint     `json:"user_id"`
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
	Domain      string   `json:"domain"`
	jwt.RegisteredClaims
}

func New(secret string, expiresIn time.Duration) *Service {
	return &Service{secret: []byte(secret), expiresIn: expiresIn}
}

func (s *Service) Generate(userID uint, email, domain string, permissions, roles []string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.expiresIn)
	claims := Claims{
		UserID:      userID,
		Email:       email,
		Permissions: permissions,
		Roles:       roles,
		Domain:      domain,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	val, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return val, expiresAt, nil
}

func (s *Service) Validate(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
