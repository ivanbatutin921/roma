package models

import (
	"time"
)

// User represents user model
type User struct {
	ID            string    `json:"id"`
	Login         string    `json:"login"`
	Email         string    `json:"email"`
	Password      string    `json:"-"`
	ProfileImage  string    `json:"profile_image,omitempty"`
	ProfileBanner string    `json:"profile_banner,omitempty"`
	Description   string    `json:"description,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserLogin представляет данные для авторизации
type UserLogin struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserRegister представляет данные для регистрации
type UserRegister struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserResponse представляет данные пользователя для ответа
type UserResponse struct {
	ID            string    `json:"id"`
	Login         string    `json:"login"`
	Email         string    `json:"email"`
	ProfileImage  string    `json:"profile_image,omitempty"`
	ProfileBanner string    `json:"profile_banner,omitempty"`
	Description   string    `json:"description,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// TokenResponse представляет ответ с токеном авторизации
type TokenResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
