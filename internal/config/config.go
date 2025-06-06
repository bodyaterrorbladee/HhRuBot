// Настройка конфига
package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	TelegramBotToken string
	TelegramChatId   string
	RedisAddr        string
	RedisPassword    string
	RedisDB          int
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env Файл не найден")
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0

	return &Config{
		TelegramBotToken: botToken,
		TelegramChatId:   chatID,
		RedisAddr:        redisAddr,
		RedisPassword:    redisPassword,
		RedisDB:          redisDB,
	}
}
