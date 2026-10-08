package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port              string
	Env               string
	DBType            string
	DBDSN             string
	SessionSecret     string
	AdminUser         string
	AdminPass         string
	TelegramBotToken  string
	TelegramChatID    string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite"
	}

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "tahoorian.db"
	}

	sessionSecret := os.Getenv("SESSION_SECRET")
	if sessionSecret == "" {
		sessionSecret = "tahoorian-secret-2026-production"
	}

	adminUser := os.Getenv("ADMIN_USER")
	if adminUser == "" {
		adminUser = "admin"
	}

	adminPass := os.Getenv("ADMIN_PASS")
	if adminPass == "" {
		adminPass = "09152491524"
	}

	return &Config{
		Port:             port,
		Env:              env,
		DBType:           dbType,
		DBDSN:            dbDSN,
		SessionSecret:    sessionSecret,
		AdminUser:        adminUser,
		AdminPass:        adminPass,
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
	}
}

func (c *Config) PortInt() int {
	p, err := strconv.Atoi(c.Port)
	if err != nil {
		return 8080
	}
	return p
}
