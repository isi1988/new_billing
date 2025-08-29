package config

import (
	"os"
	"strconv"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Auth     AuthConfig     `yaml:"auth"` // ДОБАВЛЕНО
	Database DatabaseConfig `yaml:"database"`
	Nfcapd   NfcapdConfig   `yaml:"nfcapd"`
	Telegram TelegramConfig `yaml:"telegram"`
	SMTP     SMTPConfig     `yaml:"smtp"`
}

// ДОБАВЛЕНА СТРУКТУРА
type AuthConfig struct {
	JWTSecret string `yaml:"jwt_secret"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type NfcapdConfig struct {
	Directory    string `yaml:"directory"`
	ScanInterval string `yaml:"scan_interval"`
}

type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
	Enabled  bool   `yaml:"enabled"`
}

type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	FromName string `yaml:"from_name"`
	Enabled  bool   `yaml:"enabled"`
}

func LoadConfig(path string) (*Config, error) {
	config := &Config{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	
	// Переопределяем SMTP настройки из переменных окружения, если они заданы
	if smtpHost := os.Getenv("SMTP_HOST"); smtpHost != "" {
		config.SMTP.Host = smtpHost
	}
	if smtpPortStr := os.Getenv("SMTP_PORT"); smtpPortStr != "" {
		if smtpPort, parseErr := strconv.Atoi(smtpPortStr); parseErr == nil {
			config.SMTP.Port = smtpPort
		}
	}
	if smtpUsername := os.Getenv("SMTP_USERNAME"); smtpUsername != "" {
		config.SMTP.Username = smtpUsername
	}
	if smtpPassword := os.Getenv("SMTP_PASSWORD"); smtpPassword != "" {
		config.SMTP.Password = smtpPassword
	}
	if smtpFrom := os.Getenv("SMTP_FROM"); smtpFrom != "" {
		config.SMTP.From = smtpFrom
	}
	if smtpFromName := os.Getenv("SMTP_FROM_NAME"); smtpFromName != "" {
		config.SMTP.FromName = smtpFromName
	}
	if smtpEnabledStr := os.Getenv("SMTP_ENABLED"); smtpEnabledStr != "" {
		if smtpEnabled, parseErr := strconv.ParseBool(smtpEnabledStr); parseErr == nil {
			config.SMTP.Enabled = smtpEnabled
		}
	}
	
	return config, nil
}
