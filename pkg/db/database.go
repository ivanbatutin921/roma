package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// DB instance
var DB *sql.DB

// InitDatabase initializes the database
func InitDatabase() {
	var err error
	DB, err = sql.Open("sqlite3", "./sqlite.db")
	if err != nil {
		log.Fatalf("Ошибка открытия базы данных: %v", err)
	}

	// Проверка соединения
	err = DB.Ping()
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	// Создаем таблицы, если они не существуют
	createTables()

	log.Println("База данных успешно инициализирована")
}

// createTables создает таблицы в базе данных
func createTables() {
	// Создание таблицы пользователей
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		login TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		profile_image TEXT,
		profile_banner TEXT,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	// Создание таблицы карточек
	createCardsTable := `
	CREATE TABLE IF NOT EXISTS cards (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		user_name TEXT NOT NULL,
		image TEXT,
		title TEXT NOT NULL,
		description TEXT,
		text TEXT,
		likes INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	// Создание таблицы лайков
	createLikesTable := `
	CREATE TABLE IF NOT EXISTS likes (
		user_id TEXT NOT NULL,
		card_id TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, card_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (card_id) REFERENCES cards(id) ON DELETE CASCADE
	);`

	// Выполнение запросов создания таблиц
	_, err := DB.Exec(createUsersTable)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы пользователей: %v", err)
	}

	_, err = DB.Exec(createCardsTable)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы карточек: %v", err)
	}

	_, err = DB.Exec(createLikesTable)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы лайков: %v", err)
	}
}
