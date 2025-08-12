package main

import (
	"log"

	"hhruBot/internal/config"
	"hhruBot/internal/hh"
	"hhruBot/internal/storage"
	"hhruBot/internal/telegram"
)

func main() {
	log.Println("⚙️ Загрузка конфигурации...")
	cfg := config.LoadConfig()

	log.Println("⚙️ Подключение к хранилищу...")
	store := storage.NewStorage(cfg)

	log.Println("⚙️ Инициализация карты городов...")
	if err := hh.InitCityMap(); err != nil {
		log.Fatalf("❌ Не удалось инициализировать карту городов: %v", err)
	}

	log.Println("⚙️ Создание и запуск Telegram-бота...")
	bot := telegram.NewBot(cfg, store)

	go bot.Start()

	log.Println("✅ Бот запущен и готов к работе.")
	select {}
}
