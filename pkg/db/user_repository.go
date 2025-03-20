package db

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/user/roma/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser создает нового пользователя в базе данных
func CreateUser(user models.UserRegister) (*models.User, error) {
	// Проверка на уникальность логина и email
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE login = ? OR email = ?)",
		user.Login, user.Email).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("пользователь с таким логином или email уже существует")
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Создание нового пользователя
	newUser := &models.User{
		ID:        uuid.NewString(),
		Login:     user.Login,
		Email:     user.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Добавление пользователя в базу данных
	_, err = DB.Exec(`
		INSERT INTO users (id, login, email, password, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		newUser.ID, newUser.Login, newUser.Email, newUser.Password,
		newUser.CreatedAt, newUser.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

// GetUserByID получает пользователя по ID
func GetUserByID(id string) (*models.User, error) {
	user := &models.User{}
	var profileImage, profileBanner, description sql.NullString

	err := DB.QueryRow(`
		SELECT id, login, email, password, profile_image, profile_banner, description, created_at, updated_at 
		FROM users WHERE id = ?`, id).Scan(
		&user.ID, &user.Login, &user.Email, &user.Password,
		&profileImage, &profileBanner, &description,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}

	// Преобразуем sql.NullString в строки
	if profileImage.Valid {
		user.ProfileImage = profileImage.String
	}
	if profileBanner.Valid {
		user.ProfileBanner = profileBanner.String
	}
	if description.Valid {
		user.Description = description.String
	}

	return user, nil
}

// GetUserByCredentials получает пользователя по логину/email и паролю
func GetUserByCredentials(login, email, password string) (*models.User, error) {
	user := &models.User{}
	var hashedPassword string
	var profileImage, profileBanner, description sql.NullString

	// Пытаемся найти пользователя по логину или email
	err := DB.QueryRow(`
		SELECT id, login, email, password, profile_image, profile_banner, description, created_at, updated_at 
		FROM users WHERE login = ? OR email = ?`, login, email).Scan(
		&user.ID, &user.Login, &user.Email, &hashedPassword,
		&profileImage, &profileBanner, &description,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}

	// Проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return nil, errors.New("неверный пароль")
	}

	// Преобразуем sql.NullString в строки
	if profileImage.Valid {
		user.ProfileImage = profileImage.String
	}
	if profileBanner.Valid {
		user.ProfileBanner = profileBanner.String
	}
	if description.Valid {
		user.Description = description.String
	}

	user.Password = hashedPassword
	return user, nil
}

// UpdateUserProfile обновляет профиль пользователя
func UpdateUserProfile(userID string, profileImage, profileBanner, description string) error {
	_, err := DB.Exec(`
		UPDATE users 
		SET profile_image = ?, profile_banner = ?, description = ?, updated_at = ? 
		WHERE id = ?`,
		profileImage, profileBanner, description, time.Now(), userID)
	return err
}

// ToUserResponse преобразует User в UserResponse
func ToUserResponse(user *models.User) models.UserResponse {
	// Базовый URL для изображений
	baseURL := "http://localhost:3000/uploads/" // Порт берется из конфигурации сервера

	// Формируем полные URL для изображений
	var profileImage, profileBanner string
	if user.ProfileImage != "" {
		profileImage = baseURL + user.ProfileImage
	}
	if user.ProfileBanner != "" {
		profileBanner = baseURL + user.ProfileBanner
	}

	return models.UserResponse{
		ID:            user.ID,
		Login:         user.Login,
		Email:         user.Email,
		ProfileImage:  profileImage,
		ProfileBanner: profileBanner,
		Description:   user.Description,
		CreatedAt:     user.CreatedAt,
	}
}
