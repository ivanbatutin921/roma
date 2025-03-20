package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/user/roma/pkg/db"
	"github.com/user/roma/pkg/utils"
)

// Auth middleware для проверки JWT токена
func Auth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Получаем заголовок Authorization
		authHeader := c.Get("Authorization")

		// Проверяем наличие заголовка
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "требуется авторизация",
			})
		}

		// Проверяем формат заголовка
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "неверный формат токена",
			})
		}

		// Получаем токен
		tokenString := parts[1]

		// Проверяем токен
		userID, err := utils.VerifyJWT(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "неверный токен",
			})
		}

		// Получаем пользователя из базы данных
		user, err := db.GetUserByID(userID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "пользователь не найден",
			})
		}

		// Сохраняем пользователя в локальном хранилище для использования в обработчиках
		c.Locals("userID", user.ID)
		c.Locals("user", user)

		return c.Next()
	}
}
