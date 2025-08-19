package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Auth     AuthConfig     `yaml:"auth"` // ДОБАВЛЕНО
	Database DatabaseConfig `yaml:"database"`
	Nfcapd   NfcapdConfig   `yaml:"nfcapd"`
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
	return config, nil
}
