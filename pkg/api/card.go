package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/user/roma/pkg/db"
	"github.com/user/roma/pkg/models"
	"github.com/user/roma/pkg/utils"
)

// CreateCard создает новую карточку
func CreateCard(c *fiber.Ctx) error {
	// Получаем пользователя из локального хранилища (установленного middleware)
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "пользователь не найден",
		})
	}

	// Парсим данные карточки из формы
	title := c.FormValue("title")
	description := c.FormValue("description")
	text := c.FormValue("text")

	// Проверяем обязательные поля
	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "заголовок карточки обязателен",
		})
	}

	// Создаем карточку без изображения
	var imagePath string
	var err error

	// Получаем файл из запроса (если есть)
	file, err := c.FormFile("image")
	if err == nil && file != nil { // Если файл предоставлен
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
		tempPath := fmt.Sprintf("./temp/card_%s%s",
			utils.GenerateRandomString(8), filepath.Ext(file.Filename))

		// Сохраняем загруженный файл во временную директорию
		if err := c.SaveFile(file, tempPath); err != nil {
			log.Printf("Ошибка при сохранении файла: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "ошибка при сохранении файла",
			})
		}
		defer os.Remove(tempPath) // Удаляем временный файл после завершения

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
		filename := fmt.Sprintf("card_%s%s", utils.GenerateRandomString(8), ext)
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

		imagePath = filename
	}

	// Создаем карточку в базе данных
	cardCreate := models.CardCreate{
		Title:       title,
		Description: description,
		Text:        text,
		Image:       imagePath,
	}

	card, err := db.CreateCard(cardCreate, user.ID, user.Login)
	if err != nil {
		// В случае ошибки удаляем загруженное изображение
		if imagePath != "" {
			if err := utils.RemoveImage(imagePath); err != nil {
				log.Printf("Ошибка при удалении изображения: %v", err)
			}
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("ошибка создания карточки: %v", err),
		})
	}

	// Базовый URL для изображений
	baseURL := BaseImagesURL

	// Формируем ответ
	cardResponse := models.CardResponse{
		ID:          card.ID,
		UserID:      card.UserID,
		UserName:    card.UserName,
		Image:       getImageURL(baseURL, card.Image),
		Title:       card.Title,
		Description: card.Description,
		Text:        card.Text,
		Likes:       card.Likes,
		IsLiked:     false,
		CreatedAt:   card.CreatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(cardResponse)
}

// GetCards получает все карточки с пагинацией
func GetCards(c *fiber.Ctx) error {
	// Получаем параметры пагинации из запроса
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "4"))
	if err != nil || limit < 1 {
		limit = 4
	}

	// Получаем текущего пользователя, если авторизован
	var currentUserID string
	if user, ok := c.Locals("user").(*models.User); ok {
		currentUserID = user.ID
	}

	// Получаем карточки из базы данных
	response, err := db.GetAllCards(page, limit, currentUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка получения карточек",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetUserCards получает карточки пользователя с пагинацией
func GetUserCards(c *fiber.Ctx) error {
	// Получаем ID пользователя из URL
	userID := c.Params("userId")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID пользователя не указан",
		})
	}

	// Получаем параметры пагинации из запроса
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.Query("limit", "4"))
	if err != nil || limit < 1 {
		limit = 4
	}

	// Получаем текущего пользователя, если авторизован
	var currentUserID string
	if user, ok := c.Locals("user").(*models.User); ok {
		currentUserID = user.ID
	}

	// Получаем карточки пользователя из базы данных
	response, err := db.GetUserCards(userID, page, limit, currentUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка получения карточек пользователя",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetCard получает карточку по ID
func GetCard(c *fiber.Ctx) error {
	// Получаем ID карточки из URL
	cardID := c.Params("cardId")
	if cardID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID карточки не указан",
		})
	}

	// Получаем карточку из базы данных
	card, err := db.GetCardByID(cardID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "карточка не найдена",
		})
	}

	// Проверяем, лайкнул ли текущий пользователь карточку
	isLiked := false
	if user, ok := c.Locals("user").(*models.User); ok {
		for _, likedBy := range card.LikedBy {
			if likedBy == user.ID {
				isLiked = true
				break
			}
		}
	}

	// Базовый URL для изображений
	baseURL := BaseImagesURL

	// Формируем ответ
	cardResponse := models.CardResponse{
		ID:          card.ID,
		UserID:      card.UserID,
		UserName:    card.UserName,
		Image:       getImageURL(baseURL, card.Image),
		Title:       card.Title,
		Description: card.Description,
		Text:        card.Text,
		Likes:       card.Likes,
		IsLiked:     isLiked,
		CreatedAt:   card.CreatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(cardResponse)
}

// UpdateCard обновляет карточку
func UpdateCard(c *fiber.Ctx) error {
	// Получаем пользователя из локального хранилища (установленного middleware)
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "пользователь не найден",
		})
	}

	// Получаем ID карточки из URL
	cardID := c.Params("cardId")
	if cardID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID карточки не указан",
		})
	}

	// Получаем карточку из базы данных
	card, err := db.GetCardByID(cardID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "карточка не найдена",
		})
	}

	// Проверяем, что карточка принадлежит текущему пользователю
	if card.UserID != user.ID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "нет прав на редактирование этой карточки",
		})
	}

	// Парсим данные для обновления карточки из формы
	title := c.FormValue("title", card.Title)
	description := c.FormValue("description", card.Description)
	text := c.FormValue("text", card.Text)

	// Проверяем обязательные поля
	if title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "заголовок карточки обязателен",
		})
	}

	// Создаем объект с данными для обновления
	cardUpdate := models.CardUpdate{
		Title:       title,
		Description: description,
		Text:        text,
		Image:       card.Image, // По умолчанию оставляем текущее изображение
	}

	// Если есть новое изображение, сохраняем его и удаляем старое
	var newImagePath string

	// Получаем файл из запроса (если есть)
	file, err := c.FormFile("image")
	if err == nil && file != nil { // Если файл предоставлен
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
		tempPath := fmt.Sprintf("./temp/card_%s%s",
			utils.GenerateRandomString(8), filepath.Ext(file.Filename))

		// Сохраняем загруженный файл во временную директорию
		if err := c.SaveFile(file, tempPath); err != nil {
			log.Printf("Ошибка при сохранении файла: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "ошибка при сохранении файла",
			})
		}
		defer os.Remove(tempPath) // Удаляем временный файл после завершения

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
		newImagePath = fmt.Sprintf("card_%s%s", utils.GenerateRandomString(8), ext)
		savePath := filepath.Join(utils.ImageDir, newImagePath)

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

		// Если было старое изображение, удаляем его
		if card.Image != "" && card.Image != newImagePath {
			if err := utils.RemoveImage(card.Image); err != nil {
				log.Printf("Ошибка при удалении старого изображения: %v", err)
				// Продолжаем работу, не критическая ошибка
			}
		}

		// Обновляем путь к изображению
		cardUpdate.Image = newImagePath
	} else {
		// Если запрос содержит явное указание удалить изображение
		removeImage := c.FormValue("remove_image", "false")
		if removeImage == "true" {
			// Удаляем старое изображение, если оно было
			if card.Image != "" {
				if err := utils.RemoveImage(card.Image); err != nil {
					log.Printf("Ошибка при удалении старого изображения: %v", err)
					// Продолжаем работу, не критическая ошибка
				}
			}
			cardUpdate.Image = "" // Очищаем путь к изображению
		}
	}

	// Обновляем карточку в базе данных
	err = db.UpdateCard(cardID, cardUpdate)
	if err != nil {
		// В случае ошибки удаляем новое изображение, если оно было загружено
		if newImagePath != "" {
			if err := utils.RemoveImage(newImagePath); err != nil {
				log.Printf("Ошибка при удалении изображения: %v", err)
			}
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("ошибка обновления карточки: %v", err),
		})
	}

	// Получаем обновленную карточку
	updatedCard, err := db.GetCardByID(cardID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("ошибка получения обновленной карточки: %v", err),
		})
	}

	// Базовый URL для изображений
	baseURL := BaseImagesURL

	// Формируем ответ
	cardResponse := models.CardResponse{
		ID:          updatedCard.ID,
		UserID:      updatedCard.UserID,
		UserName:    updatedCard.UserName,
		Image:       getImageURL(baseURL, updatedCard.Image),
		Title:       updatedCard.Title,
		Description: updatedCard.Description,
		Text:        updatedCard.Text,
		Likes:       updatedCard.Likes,
		IsLiked:     false, // По умолчанию false, обновляется ниже
		CreatedAt:   updatedCard.CreatedAt,
	}

	// Проверяем, лайкнул ли текущий пользователь карточку
	for _, likedBy := range updatedCard.LikedBy {
		if likedBy == user.ID {
			cardResponse.IsLiked = true
			break
		}
	}

	return c.Status(fiber.StatusOK).JSON(cardResponse)
}

// DeleteCard удаляет карточку
func DeleteCard(c *fiber.Ctx) error {
	// Получаем пользователя из локального хранилища (установленного middleware)
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "пользователь не найден",
		})
	}

	// Получаем ID карточки из URL
	cardID := c.Params("cardId")
	if cardID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID карточки не указан",
		})
	}

	// Получаем карточку из базы данных
	card, err := db.GetCardByID(cardID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "карточка не найдена",
		})
	}

	// Проверяем, что карточка принадлежит текущему пользователю
	if card.UserID != user.ID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "нет прав на удаление этой карточки",
		})
	}

	// Удаляем изображение карточки, если оно есть
	if card.Image != "" {
		utils.RemoveImage(card.Image)
	}

	// Удаляем карточку из базы данных
	err = db.DeleteCard(cardID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка удаления карточки",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "карточка успешно удалена",
	})
}

// LikeCard добавляет лайк карточке
func LikeCard(c *fiber.Ctx) error {
	// Получаем пользователя из локального хранилища (установленного middleware)
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "пользователь не найден",
		})
	}

	// Получаем ID карточки из URL
	cardID := c.Params("cardId")
	if cardID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID карточки не указан",
		})
	}

	// Проверяем существование карточки
	_, err := db.GetCardByID(cardID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "карточка не найдена",
		})
	}

	// Добавляем лайк
	err = db.LikeCard(cardID, user.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Получаем обновленную карточку
	updatedCard, err := db.GetCardByID(cardID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка получения обновленной карточки",
		})
	}

	// Формируем ответ
	cardResponse := models.CardResponse{
		ID:          updatedCard.ID,
		UserID:      updatedCard.UserID,
		UserName:    updatedCard.UserName,
		Image:       updatedCard.Image,
		Title:       updatedCard.Title,
		Description: updatedCard.Description,
		Text:        updatedCard.Text,
		Likes:       updatedCard.Likes,
		IsLiked:     true, // Мы только что лайкнули
		CreatedAt:   updatedCard.CreatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(cardResponse)
}

// UnlikeCard удаляет лайк с карточки
func UnlikeCard(c *fiber.Ctx) error {
	// Получаем пользователя из локального хранилища (установленного middleware)
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "пользователь не найден",
		})
	}

	// Получаем ID карточки из URL
	cardID := c.Params("cardId")
	if cardID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID карточки не указан",
		})
	}

	// Проверяем существование карточки
	_, err := db.GetCardByID(cardID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "карточка не найдена",
		})
	}

	// Удаляем лайк
	err = db.UnlikeCard(cardID, user.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Получаем обновленную карточку
	updatedCard, err := db.GetCardByID(cardID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка получения обновленной карточки",
		})
	}

	// Формируем ответ
	cardResponse := models.CardResponse{
		ID:          updatedCard.ID,
		UserID:      updatedCard.UserID,
		UserName:    updatedCard.UserName,
		Image:       updatedCard.Image,
		Title:       updatedCard.Title,
		Description: updatedCard.Description,
		Text:        updatedCard.Text,
		Likes:       updatedCard.Likes,
		IsLiked:     false, // Мы только что убрали лайк
		CreatedAt:   updatedCard.CreatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(cardResponse)
}

// Вспомогательная функция для формирования URL изображения
func getImageURL(baseURL, imagePath string) string {
	if imagePath == "" {
		return ""
	}
	return baseURL + imagePath
}
