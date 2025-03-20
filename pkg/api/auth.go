package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/user/roma/pkg/db"
	"github.com/user/roma/pkg/models"
	"github.com/user/roma/pkg/utils"
)

// Register обработчик для регистрации пользователя
func Register(c *fiber.Ctx) error {
	// Парсим данные пользователя из тела запроса
	var userRegister models.UserRegister
	if err := c.BodyParser(&userRegister); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "неверный формат данных",
		})
	}

	// Проверяем обязательные поля
	if userRegister.Login == "" || userRegister.Email == "" || userRegister.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "все поля обязательны для заполнения",
		})
	}

	// Создаем пользователя в базе данных
	user, err := db.CreateUser(userRegister)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Генерируем JWT токен
	token, err := utils.GenerateJWT(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка создания токена",
		})
	}

	// Возвращаем токен и данные пользователя
	return c.Status(fiber.StatusCreated).JSON(models.TokenResponse{
		Token: token,
		User:  db.ToUserResponse(user),
	})
}

// Login обработчик для авторизации пользователя
func Login(c *fiber.Ctx) error {
	// Парсим данные пользователя из тела запроса
	var userLogin models.UserLogin
	if err := c.BodyParser(&userLogin); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "неверный формат данных",
		})
	}

	// Проверяем обязательные поля (login или email, и пароль)
	if (userLogin.Login == "" && userLogin.Email == "") || userLogin.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "логин/email и пароль обязательны для заполнения",
		})
	}

	// Получаем пользователя по учетным данным
	user, err := db.GetUserByCredentials(userLogin.Login, userLogin.Email, userLogin.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Генерируем JWT токен
	token, err := utils.GenerateJWT(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ошибка создания токена",
		})
	}

	// Возвращаем токен и данные пользователя
	return c.Status(fiber.StatusOK).JSON(models.TokenResponse{
		Token: token,
		User:  db.ToUserResponse(user),
	})
}

// Me обработчик для получения информации о текущем пользователе
func Me(c *fiber.Ctx) error {
	// Получаем пользователя из локального хранилища (установленного middleware)
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "пользователь не найден",
		})
	}

	// Возвращаем данные пользователя
	return c.Status(fiber.StatusOK).JSON(db.ToUserResponse(user))
}
