package models

import (
	"time"
)

// Card представляет карточку пользователя
type Card struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	Image       string    `json:"image"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Text        string    `json:"text"`
	Likes       int       `json:"likes"`
	LikedBy     []string  `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CardCreate представляет данные для создания карточки
type CardCreate struct {
	Image       string `json:"image"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Text        string `json:"text"`
}

// CardUpdate представляет данные для обновления карточки
type CardUpdate struct {
	Image       string `json:"image"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Text        string `json:"text"`
}

// CardResponse представляет карточку для ответа
type CardResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	Image       string    `json:"image"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Text        string    `json:"text"`
	Likes       int       `json:"likes"`
	IsLiked     bool      `json:"is_liked"`
	CreatedAt   time.Time `json:"created_at"`
}

// PaginationResponse представляет ответ с пагинацией
type PaginationResponse struct {
	Cards      []CardResponse `json:"cards"`
	Page       int            `json:"page"`
	TotalPages int            `json:"total_pages"`
	TotalCards int            `json:"total_cards"`
}
