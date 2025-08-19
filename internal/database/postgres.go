package database

import (
	"fmt"
	"log"
	"new-billing/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Драйвер PostgreSQL, импортируется для регистрации
	"golang.org/x/crypto/bcrypt"
)

// Connect устанавливает соединение с базой данных PostgreSQL, используя конфигурацию.
// В случае ошибки приложение завершит работу.
func Connect(cfg config.DatabaseConfig) *sqlx.DB {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		log.Fatalf("FATAL: Failed to connect to database: %v", err)
	}
	return db
}

// Migrate выполняет миграцию схемы базы данных.
// Она создает все необходимые таблицы, если они еще не существуют.
// Также создает пользователя-администратора по умолчанию при первом запуске.
func Migrate(db *sqlx.DB) {
	schema := `
	-- Таблицы для модуля NetFlow
	CREATE TABLE IF NOT EXISTS processed_files (
		id SERIAL PRIMARY KEY,
		file_name VARCHAR(255) UNIQUE NOT NULL,
		processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS flows (
		id BIGSERIAL PRIMARY KEY,
		"timestamp" TIMESTAMP WITH TIME ZONE NOT NULL,
		src_ip VARCHAR(45) NOT NULL,
		dst_ip VARCHAR(45) NOT NULL,
		src_port INT NOT NULL,
		dst_port INT NOT NULL,
		protocol INT NOT NULL,
		packets BIGINT NOT NULL,
		bytes BIGINT NOT NULL
	);
	
	-- Таблицы для модуля биллинга
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		role VARCHAR(20) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE TABLE IF NOT EXISTS clients (
		id SERIAL PRIMARY KEY,
		client_type VARCHAR(20) NOT NULL,
		email VARCHAR(100),
		phone VARCHAR(20),
		-- Поля для физ. лиц
		first_name VARCHAR(100),
		last_name VARCHAR(100),
		patronymic VARCHAR(100),
		passport_number VARCHAR(50),
		passport_issued_by TEXT,
		passport_issue_date DATE,
		registration_address TEXT,
		birth_date DATE,
		-- Поля для юр. лиц
		inn VARCHAR(12) UNIQUE,
		kpp VARCHAR(9),
		full_name TEXT,
		short_name VARCHAR(255),
		ogrn VARCHAR(15) UNIQUE,
		ogrn_date DATE,
		legal_address TEXT,
		actual_address TEXT,
		bank_details TEXT,
		ceo VARCHAR(255),
		accountant VARCHAR(255)
	);

	CREATE TABLE IF NOT EXISTS equipment (
		id SERIAL PRIMARY KEY,
		model VARCHAR(100) NOT NULL,
		description TEXT,
		mac_address VARCHAR(17) UNIQUE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS tariffs (
		id SERIAL PRIMARY KEY,
		"name" VARCHAR(255) NOT NULL,
		is_archived BOOLEAN DEFAULT FALSE,
		payment_type VARCHAR(20) NOT NULL,
		is_for_individuals BOOLEAN NOT NULL,
		max_speed_in INT NOT NULL,
		max_speed_out INT NOT NULL,
		max_traffic_in BIGINT DEFAULT 0,
		max_traffic_out BIGINT DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS contracts (
		id SERIAL PRIMARY KEY,
		client_id INT NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
		"number" VARCHAR(50) UNIQUE NOT NULL,
		sign_date DATE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS connections (
		id SERIAL PRIMARY KEY,
		equipment_id INT NOT NULL REFERENCES equipment(id),
		contract_id INT NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
		address TEXT NOT NULL,
		connection_type VARCHAR(50),
		tariff_id INT NOT NULL REFERENCES tariffs(id),
		ip_address VARCHAR(45) UNIQUE NOT NULL,
		mask INT NOT NULL
	);
	`
	// MustExec выполняет запрос и паникует в случае ошибки.
	// Это допустимо при старте приложения, так как без корректной схемы оно не может работать.
	db.MustExec(schema)

	// Проверяем и создаем пользователя 'admin' с паролем 'admin', если его нет.
	var userCount int
	err := db.Get(&userCount, "SELECT count(*) FROM users WHERE username='admin'")
	if err == nil && userCount == 0 {
		log.Println("Default 'admin' user not found, creating one...")
		password := "admin" // В реальном приложении пароль лучше получать из переменных окружения
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Could not hash password for default admin user: %v", err)
		}

		_, err = db.Exec("INSERT INTO users (username, password_hash, role) VALUES ('admin', $1, 'admin')", string(hashedPassword))
		if err != nil {
			log.Fatalf("Could not insert default admin user: %v", err)
		}
		log.Println("Default 'admin' user created successfully with password 'admin'.")
	}
}
