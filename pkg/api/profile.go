package api

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/user/roma/pkg/db"
	"github.com/user/roma/pkg/models"
	"github.com/user/roma/pkg/utils"
)

// Константа для базового URL изображений
const BaseImagesURL = "http://localhost:4000/uploads/"

// UpdateProfile обновляет профиль пользователя
func UpdateProfile(c *fiber.Ctx) error {
	// Получаем пользователя из локального хранилища (установленного middleware)
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "пользователь не найден",
		})
	}

	// Парсим данные профиля из тела запроса
	var profileData struct {
		Description string `json:"description"`
	}
	if err := c.BodyParser(&profileData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "неверный формат данных",
		})
	}

	// Обновляем профиль в базе данных (оставляем существующие изображения без изменений)
	err := db.UpdateUserProfile(user.ID, user.ProfileImage, user.ProfileBanner, profileData.Description)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("ошибка обновления профиля: %v", err),
		})
	}

	// Получаем обновленные данные пользователя
	updatedUser, err := db.GetUserByID(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("ошибка получения обновленных данных профиля: %v", err),
		})
	}

	// Возвращаем обновленные данные пользователя
	return c.Status(fiber.StatusOK).JSON(db.ToUserResponse(updatedUser))
}

// UploadProfileImage загружает изображение профиля
func UploadProfileImage(c *fiber.Ctx) error {
	// Получаем пользователя из локального хранилища (установленного middleware)
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "пользователь не найден",
		})
	}

	// Получаем файл из запроса
	file, err := c.FormFile("image")
	if err != nil {
		log.Printf("Ошибка при получении файла: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("не удалось получить файл: %v", err),
		})
	}

	// Проверяем размер файла
	if file.Size > utils.MaxImageSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "размер изображения превышает максимально допустимый",
		})
	}

	// Создаем директорию для временных файлов, если она не существует
	if err := os.MkdirAll("./temp", 0755); err != nil {
		log.Printf("Ошибка при создании директории для временных файлов: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при обработке файла",
		})
	}

	// Создаем уникальное имя для временного файла
	tempPath := fmt.Sprintf("./temp/profile_%s%s",
		utils.GenerateRandomString(8), filepath.Ext(file.Filename))

	// Сохраняем загруженный файл во временную директорию
	if err := c.SaveFile(file, tempPath); err != nil {
		log.Printf("Ошибка при сохранении файла: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при сохранении файла",
		})
	}
	defer os.Remove(tempPath) // Удаляем временный файл после завершения

	// Если у пользователя уже есть изображение профиля, удаляем его
	if user.ProfileImage != "" {
		if err := utils.RemoveImage(user.ProfileImage); err != nil {
			log.Printf("Ошибка при удалении старого изображения профиля: %v", err)
			// продолжаем работу, не критическая ошибка
		}
	}

	// Читаем файл и сохраняем его с правильным именем в директории uploads
	fileData, err := os.ReadFile(tempPath)
	if err != nil {
		log.Printf("Ошибка при чтении временного файла: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при обработке файла",
		})
	}

	// Определяем тип контента
	contentType := http.DetectContentType(fileData)
	var ext string
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("неподдерживаемый формат изображения: %s", contentType),
		})
	}

	// Генерируем имя файла и сохраняем его
	filename := fmt.Sprintf("profile_%s%s", utils.GenerateRandomString(8), ext)
	savePath := filepath.Join(utils.ImageDir, filename)

	// Создаем директорию для загрузок, если она не существует
	if err := os.MkdirAll(utils.ImageDir, 0755); err != nil {
		log.Printf("Ошибка при создании директории для загрузок: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при сохранении файла",
		})
	}

	// Копируем файл
	if err := os.WriteFile(savePath, fileData, 0644); err != nil {
		log.Printf("Ошибка при сохранении файла: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при сохранении файла",
		})
	}

	// Обновляем профиль в базе данных с новым изображением
	err = db.UpdateUserProfile(user.ID, filename, user.ProfileBanner, user.Description)
	if err != nil {
		// Если произошла ошибка, удаляем загруженное изображение
		if err := utils.RemoveImage(filename); err != nil {
			log.Printf("Ошибка при удалении временного изображения: %v", err)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("ошибка обновления профиля: %v", err),
		})
	}

	// Получаем обновленные данные пользователя
	updatedUser, err := db.GetUserByID(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("ошибка получения обновленных данных профиля: %v", err),
		})
	}

	// Возвращаем обновленные данные пользователя
	return c.Status(fiber.StatusOK).JSON(db.ToUserResponse(updatedUser))
}

// UploadProfileBanner загружает баннер профиля
func UploadProfileBanner(c *fiber.Ctx) error {
	// Получаем пользователя из локального хранилища (установленного middleware)
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "пользователь не найден",
		})
	}

	// Получаем файл из запроса
	file, err := c.FormFile("image")
	if err != nil {
		log.Printf("Ошибка при получении файла: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("не удалось получить файл: %v", err),
		})
	}

	// Проверяем размер файла
	if file.Size > utils.MaxImageSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "размер изображения превышает максимально допустимый",
		})
	}

	// Создаем директорию для временных файлов, если она не существует
	if err := os.MkdirAll("./temp", 0755); err != nil {
		log.Printf("Ошибка при создании директории для временных файлов: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при обработке файла",
		})
	}

	// Создаем уникальное имя для временного файла
	tempPath := fmt.Sprintf("./temp/banner_%s%s",
		utils.GenerateRandomString(8), filepath.Ext(file.Filename))

	// Сохраняем загруженный файл во временную директорию
	if err := c.SaveFile(file, tempPath); err != nil {
		log.Printf("Ошибка при сохранении файла: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при сохранении файла",
		})
	}
	defer os.Remove(tempPath) // Удаляем временный файл после завершения

	// Если у пользователя уже есть баннер профиля, удаляем его
	if user.ProfileBanner != "" {
		if err := utils.RemoveImage(user.ProfileBanner); err != nil {
			log.Printf("Ошибка при удалении старого баннера профиля: %v", err)
			// продолжаем работу, не критическая ошибка
		}
	}

	// Читаем файл и проверяем размеры изображения
	fileData, err := os.ReadFile(tempPath)
	if err != nil {
		log.Printf("Ошибка при чтении временного файла: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при обработке файла",
		})
	}

	// Проверка размеров изображения для баннера
	img, _, err := image.Decode(bytes.NewReader(fileData))
	if err != nil {
		log.Printf("Ошибка при декодировании изображения: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "невозможно декодировать изображение",
		})
	}

	bounds := img.Bounds()
	if bounds.Dx() > utils.MaxBannerWidth || bounds.Dy() > utils.MaxBannerHeight {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("размер баннера превышает максимально допустимый (%dx%d)",
				utils.MaxBannerWidth, utils.MaxBannerHeight),
		})
	}

	// Определяем тип контента
	contentType := http.DetectContentType(fileData)
	var ext string
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("неподдерживаемый формат изображения: %s", contentType),
		})
	}

	// Генерируем имя файла и сохраняем его
	filename := fmt.Sprintf("banner_%s%s", utils.GenerateRandomString(8), ext)
	savePath := filepath.Join(utils.ImageDir, filename)

	// Создаем директорию для загрузок, если она не существует
	if err := os.MkdirAll(utils.ImageDir, 0755); err != nil {
		log.Printf("Ошибка при создании директории для загрузок: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при сохранении файла",
		})
	}

	// Копируем файл
	if err := os.WriteFile(savePath, fileData, 0644); err != nil {
		log.Printf("Ошибка при сохранении файла: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка при сохранении файла",
		})
	}

	// Обновляем профиль в базе данных с новым баннером
	err = db.UpdateUserProfile(user.ID, user.ProfileImage, filename, user.Description)
	if err != nil {
		// Если произошла ошибка, удаляем загруженное изображение
		if err := utils.RemoveImage(filename); err != nil {
			log.Printf("Ошибка при удалении временного изображения: %v", err)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("ошибка обновления профиля: %v", err),
		})
	}

	// Получаем обновленные данные пользователя
	updatedUser, err := db.GetUserByID(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("ошибка получения обновленных данных профиля: %v", err),
		})
	}

	// Возвращаем обновленные данные пользователя
	return c.Status(fiber.StatusOK).JSON(db.ToUserResponse(updatedUser))
}
