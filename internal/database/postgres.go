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
		is_blocked BOOLEAN DEFAULT FALSE,
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
		bank_name VARCHAR(255),
		bank_account VARCHAR(30),
		bank_bik VARCHAR(9),
		bank_correspondent VARCHAR(30),
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
		sign_date DATE NOT NULL,
		is_blocked BOOLEAN DEFAULT FALSE
	);

	CREATE TABLE IF NOT EXISTS connections (
		id SERIAL PRIMARY KEY,
		equipment_id INT NOT NULL REFERENCES equipment(id),
		contract_id INT NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
		address TEXT NOT NULL,
		connection_type VARCHAR(50),
		tariff_id INT NOT NULL REFERENCES tariffs(id),
		ip_address VARCHAR(45) NOT NULL,
		mask INT NOT NULL,
		is_blocked BOOLEAN DEFAULT FALSE
	);

	CREATE TABLE IF NOT EXISTS traffic (
		id SERIAL PRIMARY KEY,
		connection_id INTEGER NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
		client_id INTEGER NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
		timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
		bytes_in BIGINT NOT NULL DEFAULT 0,
		bytes_out BIGINT NOT NULL DEFAULT 0,
		packets_in BIGINT NOT NULL DEFAULT 0,
		packets_out BIGINT NOT NULL DEFAULT 0,
		created_at TIMESTAMP DEFAULT NOW()
	);

	-- Создаем индексы для таблицы traffic
	CREATE INDEX IF NOT EXISTS idx_traffic_client_id ON traffic(client_id);
	CREATE INDEX IF NOT EXISTS idx_traffic_connection_id ON traffic(connection_id);
	CREATE INDEX IF NOT EXISTS idx_traffic_timestamp ON traffic(timestamp);
	CREATE INDEX IF NOT EXISTS idx_traffic_client_timestamp ON traffic(client_id, timestamp);

	-- Таблица для доработок/задач
	CREATE TABLE IF NOT EXISTS issues (
		id SERIAL PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		description TEXT NOT NULL,
		status VARCHAR(20) NOT NULL DEFAULT 'new',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		resolved_at TIMESTAMP WITH TIME ZONE,
		created_by INT NOT NULL REFERENCES users(id),
		resolved_by INT REFERENCES users(id)
	);

	-- Создаем индексы для таблицы issues
	CREATE INDEX IF NOT EXISTS idx_issues_status ON issues(status);
	CREATE INDEX IF NOT EXISTS idx_issues_created_at ON issues(created_at);
	CREATE INDEX IF NOT EXISTS idx_issues_created_by ON issues(created_by);

	-- Таблица для истории редактирования задач
	CREATE TABLE IF NOT EXISTS issue_history (
		id SERIAL PRIMARY KEY,
		issue_id INT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
		field_name VARCHAR(50) NOT NULL,
		old_value TEXT,
		new_value TEXT,
		edited_by INT NOT NULL REFERENCES users(id),
		edited_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	-- Создаем индексы для таблицы issue_history
	CREATE INDEX IF NOT EXISTS idx_issue_history_issue_id ON issue_history(issue_id);
	CREATE INDEX IF NOT EXISTS idx_issue_history_edited_at ON issue_history(edited_at);
	CREATE INDEX IF NOT EXISTS idx_issue_history_edited_by ON issue_history(edited_by);
	
	-- Таблица для комментариев к задачам (клиент-менеджер коммуникация)
	CREATE TABLE IF NOT EXISTS issue_comments (
		id SERIAL PRIMARY KEY,
		issue_id INT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
		message TEXT NOT NULL,
		author_id INT NOT NULL REFERENCES users(id),
		author_role VARCHAR(20) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	-- Создаем индексы для таблицы issue_comments
	CREATE INDEX IF NOT EXISTS idx_issue_comments_issue_id ON issue_comments(issue_id);
	CREATE INDEX IF NOT EXISTS idx_issue_comments_created_at ON issue_comments(created_at);
	CREATE INDEX IF NOT EXISTS idx_issue_comments_author_id ON issue_comments(author_id);
	
	-- Таблица для хранения соответствий IP адресов и хостов
	CREATE TABLE IF NOT EXISTS ip_hostnames (
		id SERIAL PRIMARY KEY,
		ip_address VARCHAR(45) UNIQUE NOT NULL,
		hostname VARCHAR(255) NOT NULL,
		resolved_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	
	-- Создаем индексы для таблицы ip_hostnames
	CREATE INDEX IF NOT EXISTS idx_ip_hostnames_ip ON ip_hostnames(ip_address);
	CREATE INDEX IF NOT EXISTS idx_ip_hostnames_updated_at ON ip_hostnames(updated_at);
	`
	// MustExec выполняет запрос и паникует в случае ошибки.
	// Это допустимо при старте приложения, так как без корректной схемы оно не может работать.
	db.MustExec(schema)

	// Удаляем ограничение уникальности с ip_address, если оно существует
	migrationSQL := `
		DO $$ 
		BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.table_constraints 
				WHERE constraint_name = 'connections_ip_address_key' 
				AND table_name = 'connections'
			) THEN
				ALTER TABLE connections DROP CONSTRAINT connections_ip_address_key;
			END IF;
			
			-- Добавляем поле is_blocked если его нет
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'connections' AND column_name = 'is_blocked'
			) THEN
				ALTER TABLE connections ADD COLUMN is_blocked BOOLEAN DEFAULT FALSE;
			END IF;
		END $$;
	`
	db.MustExec(migrationSQL)

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
