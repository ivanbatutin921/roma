package utils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "image/gif"

	"github.com/google/uuid"
)

const (
	// Максимальный размер изображения (5MB)
	MaxImageSize = 5 * 1024 * 1024
	// Директория для сохранения изображений
	ImageDir = "./uploads"
	// Максимальные размеры для баннера профиля
	MaxBannerWidth  = 1200
	MaxBannerHeight = 400
)

// SaveImageFromBase64 сохраняет изображение из base64 строки
func SaveImageFromBase64(base64String, prefix string) (string, error) {
	// Создаем директорию для загрузок, если ее нет
	if err := os.MkdirAll(ImageDir, 0755); err != nil {
		return "", err
	}

	// Если строка пустая
	if base64String == "" {
		return "", errors.New("пустая строка base64")
	}

	// Декодируем base64 строку
	var data []byte
	var err error

	// Проверяем наличие префикса data URL
	if idx := strings.Index(base64String, ";base64,"); idx > 0 {
		// Получаем часть после ";base64,"
		base64String = base64String[idx+8:]
	} else if strings.HasPrefix(base64String, "data:") {
		// Другой формат data URL, ищем первую запятую
		if idx := strings.Index(base64String, ","); idx > 0 {
			base64String = base64String[idx+1:]
		}
	}

	// Удаляем все пробелы и переносы строк, которые могут быть в строке
	base64String = strings.ReplaceAll(base64String, " ", "")
	base64String = strings.ReplaceAll(base64String, "\n", "")
	base64String = strings.ReplaceAll(base64String, "\r", "")
	base64String = strings.ReplaceAll(base64String, "\t", "")

	// Декодируем base64
	data, err = base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		return "", fmt.Errorf("ошибка декодирования base64: %v", err)
	}

	// Проверяем размер изображения
	if len(data) > MaxImageSize {
		return "", errors.New("размер изображения превышает максимально допустимый")
	}

	// Определяем формат изображения
	contentType := http.DetectContentType(data)
	var ext string
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	default:
		return "", fmt.Errorf("неподдерживаемый формат изображения: %s", contentType)
	}

	// Генерируем уникальное имя файла
	filename := fmt.Sprintf("%s_%s%s", prefix, uuid.NewString(), ext)
	filepath := filepath.Join(ImageDir, filename)

	// Проверяем размеры изображения, если это баннер профиля
	if prefix == "banner" {
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return "", fmt.Errorf("ошибка декодирования изображения: %v", err)
		}
		bounds := img.Bounds()
		if bounds.Dx() > MaxBannerWidth || bounds.Dy() > MaxBannerHeight {
			return "", fmt.Errorf("размер баннера превышает максимально допустимый (%dx%d)",
				MaxBannerWidth, MaxBannerHeight)
		}
	}

	// Сохраняем изображение в файл
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return "", fmt.Errorf("ошибка сохранения файла: %v", err)
	}

	// Возвращаем относительный путь к изображению
	return filename, nil
}

// RemoveImage удаляет изображение по имени файла
func RemoveImage(filename string) error {
	if filename == "" {
		return nil
	}

	filepath := filepath.Join(ImageDir, filename)
	return os.Remove(filepath)
}

// ImageExists проверяет существование изображения
func ImageExists(filename string) bool {
	if filename == "" {
		return false
	}

	filepath := filepath.Join(ImageDir, filename)
	_, err := os.Stat(filepath)
	return err == nil
}
