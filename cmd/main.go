package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/user/roma/pkg/api"
	"github.com/user/roma/pkg/db"
	"github.com/user/roma/pkg/middleware"
	"github.com/user/roma/pkg/utils"
)

func main() {
	// Загружаем переменные окружения из .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, используем переменные окружения системы")
	}

	// Инициализируем базу данных
	db.InitDatabase()

	// Создаем директорию для загрузки изображений, если ее нет
	if err := os.MkdirAll(utils.ImageDir, 0755); err != nil {
		log.Fatalf("Ошибка создания директории для загрузки изображений: %v", err)
	}

	// Создаем экземпляр Fiber
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB
	})

	// Добавляем middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Статические файлы для загрузок
	app.Static("/uploads", "./uploads")

	// Определяем маршруты API
	apiRouter := app.Group("/api")

	// Маршруты аутентификации (публичные)
	auth := apiRouter.Group("/auth")
	auth.Post("/register", api.Register)
	auth.Post("/login", api.Login)
	auth.Get("/me", middleware.Auth(), api.Me)

	// Маршруты профиля (требуют аутентификации)
	profile := apiRouter.Group("/profile", middleware.Auth())
	profile.Put("/", api.UpdateProfile)
	profile.Post("/image", api.UploadProfileImage)
	profile.Post("/banner", api.UploadProfileBanner)

	// Маршруты карточек
	cards := apiRouter.Group("/cards")
	cards.Get("/", api.GetCards)                                     // Получение всех карточек (публичный)
	cards.Get("/:cardId", api.GetCard)                               // Получение карточки по ID (публичный)
	cards.Get("/user/:userId", api.GetUserCards)                     // Получение карточек пользователя (публичный)
	cards.Post("/", middleware.Auth(), api.CreateCard)               // Создание карточки (требует аутентификации)
	cards.Put("/:cardId", middleware.Auth(), api.UpdateCard)         // Обновление карточки (требует аутентификации)
	cards.Delete("/:cardId", middleware.Auth(), api.DeleteCard)      // Удаление карточки (требует аутентификации)
	cards.Post("/:cardId/like", middleware.Auth(), api.LikeCard)     // Лайк карточки (требует аутентификации)
	cards.Delete("/:cardId/like", middleware.Auth(), api.UnlikeCard) // Удаление лайка (требует аутентификации)

	// Получаем порт из переменных окружения или используем порт по умолчанию
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Запускаем сервер
	log.Printf("Сервер запущен на порту %s", port)
	log.Fatal(app.Listen(":" + port))
}
