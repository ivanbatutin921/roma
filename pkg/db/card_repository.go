package db

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/user/roma/pkg/models"
)

// CreateCard создает новую карточку в базе данных
func CreateCard(card models.CardCreate, userID, userName string) (*models.Card, error) {
	// Создание новой карточки
	newCard := &models.Card{
		ID:          uuid.NewString(),
		UserID:      userID,
		UserName:    userName,
		Image:       card.Image,
		Title:       card.Title,
		Description: card.Description,
		Text:        card.Text,
		Likes:       0,
		LikedBy:     []string{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Добавление карточки в базу данных
	_, err := DB.Exec(`
		INSERT INTO cards (id, user_id, user_name, image, title, description, text, likes, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newCard.ID, newCard.UserID, newCard.UserName, newCard.Image, newCard.Title,
		newCard.Description, newCard.Text, newCard.Likes, newCard.CreatedAt, newCard.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return newCard, nil
}

// GetCardByID получает карточку по ID
func GetCardByID(id string) (*models.Card, error) {
	card := &models.Card{}
	err := DB.QueryRow(`
		SELECT id, user_id, user_name, image, title, description, text, likes, created_at, updated_at 
		FROM cards WHERE id = ?`, id).Scan(
		&card.ID, &card.UserID, &card.UserName, &card.Image, &card.Title,
		&card.Description, &card.Text, &card.Likes, &card.CreatedAt, &card.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("карточка не найдена")
		}
		return nil, err
	}

	// Получаем список пользователей, лайкнувших карточку
	rows, err := DB.Query("SELECT user_id FROM likes WHERE card_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	likedBy := []string{}
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		likedBy = append(likedBy, userID)
	}
	card.LikedBy = likedBy

	return card, nil
}

// GetAllCards получает все карточки с пагинацией
func GetAllCards(page, limit int, currentUserID string) (*models.PaginationResponse, error) {
	// Получаем общее количество карточек
	var totalCards int
	err := DB.QueryRow("SELECT COUNT(*) FROM cards").Scan(&totalCards)
	if err != nil {
		return nil, err
	}

	// Рассчитываем общее количество страниц
	totalPages := (totalCards + limit - 1) / limit

	// Если запрошенная страница больше общего количества страниц, возвращаем последнюю страницу
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	// Вычисляем смещение
	offset := (page - 1) * limit

	// Получаем карточки для текущей страницы
	rows, err := DB.Query(`
		SELECT id, user_id, user_name, image, title, description, text, likes, created_at, updated_at 
		FROM cards ORDER BY created_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cardResponses := []models.CardResponse{}

	// Базовый URL для изображений
	baseURL := "http://localhost:3000/uploads/" // Порт берется из конфигурации сервера

	// Формируем ответ
	for rows.Next() {
		var card models.Card
		var likedBy sql.NullString

		if err := rows.Scan(&card.ID, &card.UserID, &card.UserName, &card.Image, &card.Title,
			&card.Description, &card.Text, &card.Likes, &likedBy, &card.CreatedAt, &card.UpdatedAt); err != nil {
			return nil, err
		}

		// Парсим лайки
		if likedBy.Valid && likedBy.String != "" {
			card.LikedBy = strings.Split(likedBy.String, ",")
		}

		// Проверяем, поставил ли текущий пользователь лайк
		isLiked := false
		if currentUserID != "" {
			var exists bool
			err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND card_id = ?)",
				currentUserID, card.ID).Scan(&exists)
			if err != nil {
				return nil, err
			}
			isLiked = exists
		}

		// Формируем полный URL для изображения
		imageURL := ""
		if card.Image != "" {
			imageURL = baseURL + card.Image
		}

		cardResponse := models.CardResponse{
			ID:          card.ID,
			UserID:      card.UserID,
			UserName:    card.UserName,
			Image:       imageURL,
			Title:       card.Title,
			Description: card.Description,
			Text:        card.Text,
			Likes:       card.Likes,
			IsLiked:     isLiked,
			CreatedAt:   card.CreatedAt,
		}

		cardResponses = append(cardResponses, cardResponse)
	}

	return &models.PaginationResponse{
		Cards:      cardResponses,
		Page:       page,
		TotalPages: totalPages,
		TotalCards: totalCards,
	}, nil
}

// GetUserCards получает карточки пользователя с пагинацией
func GetUserCards(userID string, page, limit int, currentUserID string) (*models.PaginationResponse, error) {
	// Получаем общее количество карточек пользователя
	var totalCards int
	err := DB.QueryRow("SELECT COUNT(*) FROM cards WHERE user_id = ?", userID).Scan(&totalCards)
	if err != nil {
		return nil, err
	}

	// Рассчитываем общее количество страниц
	totalPages := (totalCards + limit - 1) / limit

	// Если запрошенная страница больше общего количества страниц, возвращаем последнюю страницу
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	// Вычисляем смещение
	offset := (page - 1) * limit

	// Получаем карточки пользователя для текущей страницы
	rows, err := DB.Query(`
		SELECT id, user_id, user_name, image, title, description, text, likes, created_at, updated_at 
		FROM cards WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cardResponses := []models.CardResponse{}

	// Базовый URL для изображений
	baseURL := "http://localhost:3000/uploads/" // Порт берется из конфигурации сервера

	// Формируем ответ
	for rows.Next() {
		var card models.Card
		var likedBy sql.NullString

		if err := rows.Scan(&card.ID, &card.UserID, &card.UserName, &card.Image, &card.Title,
			&card.Description, &card.Text, &card.Likes, &likedBy, &card.CreatedAt, &card.UpdatedAt); err != nil {
			return nil, err
		}

		// Парсим лайки
		if likedBy.Valid && likedBy.String != "" {
			card.LikedBy = strings.Split(likedBy.String, ",")
		}

		// Проверяем, поставил ли текущий пользователь лайк
		isLiked := false
		if currentUserID != "" {
			var exists bool
			err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND card_id = ?)",
				currentUserID, card.ID).Scan(&exists)
			if err != nil {
				return nil, err
			}
			isLiked = exists
		}

		// Формируем полный URL для изображения
		imageURL := ""
		if card.Image != "" {
			imageURL = baseURL + card.Image
		}

		cardResponse := models.CardResponse{
			ID:          card.ID,
			UserID:      card.UserID,
			UserName:    card.UserName,
			Image:       imageURL,
			Title:       card.Title,
			Description: card.Description,
			Text:        card.Text,
			Likes:       card.Likes,
			IsLiked:     isLiked,
			CreatedAt:   card.CreatedAt,
		}

		cardResponses = append(cardResponses, cardResponse)
	}

	return &models.PaginationResponse{
		Cards:      cardResponses,
		Page:       page,
		TotalPages: totalPages,
		TotalCards: totalCards,
	}, nil
}

// UpdateCard обновляет карточку
func UpdateCard(id string, card models.CardUpdate) error {
	_, err := DB.Exec(`
		UPDATE cards 
		SET image = ?, title = ?, description = ?, text = ?, updated_at = ? 
		WHERE id = ?`,
		card.Image, card.Title, card.Description, card.Text, time.Now(), id)
	return err
}

// DeleteCard удаляет карточку
func DeleteCard(id string) error {
	// Удаляем все лайки карточки
	_, err := DB.Exec("DELETE FROM likes WHERE card_id = ?", id)
	if err != nil {
		return err
	}

	// Удаляем карточку
	_, err = DB.Exec("DELETE FROM cards WHERE id = ?", id)
	return err
}

// LikeCard добавляет лайк карточке
func LikeCard(cardID, userID string) error {
	// Начинаем транзакцию
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Проверяем, лайкнул ли уже пользователь карточку
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND card_id = ?)",
		userID, cardID).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("пользователь уже лайкнул эту карточку")
	}

	// Добавляем запись о лайке
	_, err = tx.Exec("INSERT INTO likes (user_id, card_id, created_at) VALUES (?, ?, ?)",
		userID, cardID, time.Now())
	if err != nil {
		return err
	}

	// Увеличиваем счетчик лайков
	_, err = tx.Exec("UPDATE cards SET likes = likes + 1 WHERE id = ?", cardID)
	if err != nil {
		return err
	}

	// Фиксируем транзакцию
	return tx.Commit()
}

// UnlikeCard удаляет лайк с карточки
func UnlikeCard(cardID, userID string) error {
	// Начинаем транзакцию
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Проверяем, лайкнул ли пользователь карточку
	var exists bool
	err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE user_id = ? AND card_id = ?)",
		userID, cardID).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("пользователь не лайкал эту карточку")
	}

	// Удаляем запись о лайке
	_, err = tx.Exec("DELETE FROM likes WHERE user_id = ? AND card_id = ?", userID, cardID)
	if err != nil {
		return err
	}

	// Уменьшаем счетчик лайков
	_, err = tx.Exec("UPDATE cards SET likes = likes - 1 WHERE id = ?", cardID)
	if err != nil {
		return err
	}

	// Фиксируем транзакцию
	return tx.Commit()
}
