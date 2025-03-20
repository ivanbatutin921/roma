package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// GenerateRandomString генерирует случайную строку указанной длины
func GenerateRandomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// SanitizeUsername удаляет запрещенные символы из имени пользователя
// и приводит его к нижнему регистру
func SanitizeUsername(username string) string {
	// Удаляем все символы, кроме букв, цифр и подчеркиваний
	sanitized := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, username)

	// Приводим к нижнему регистру
	return strings.ToLower(sanitized)
}
