package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/user/roma/pkg/models"
)

// GenerateJWT генерирует JWT токен для пользователя
func GenerateJWT(user *models.User) (string, error) {
	secretKey := []byte(getSecretKey())

	// Создаем токен со временем истечения через 24 часа
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"login": user.Login,
		"email": user.Email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	})

	// Подписываем токен секретным ключом
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyJWT проверяет JWT токен и возвращает ID пользователя
func VerifyJWT(tokenString string) (string, error) {
	secretKey := []byte(getSecretKey())

	// Парсим токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неверный метод подписи токена")
		}
		return secretKey, nil
	})

	if err != nil {
		return "", err
	}

	// Проверяем валидность токена
	if !token.Valid {
		return "", errors.New("невалидный токен")
	}

	// Получаем claims из токена
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("невозможно получить claims из токена")
	}

	// Получаем ID пользователя из claims
	userID, ok := claims["id"].(string)
	if !ok {
		return "", errors.New("невозможно получить ID пользователя из токена")
	}

	return userID, nil
}

// getSecretKey возвращает секретный ключ для JWT
func getSecretKey() string {
	key := os.Getenv("JWT_SECRET_KEY")
	if key == "" {
		// Если ключ не задан в переменных окружения, используем дефолтный
		return "your_super_secret_key_here"
	}
	return key
}
