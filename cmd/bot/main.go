package main

import (
	"log"

	"hhruBot/internal/config"
	"hhruBot/internal/hh"
	"hhruBot/internal/storage"
	"hhruBot/internal/telegram"
)

func main() {
	// Инициализация конфигурации
	cfg := config.LoadConfig()

	// Подключение к хранилищу
	store := storage.NewStorage(cfg)

	// Инициализация Telegram-бота
	bot := telegram.NewBot(cfg, store)

	// Инициализация клиента HH.ru
	hhClient := hh.NewClient()

	// Запуск бота в отдельной горутине
	go bot.Start()

	// Получение всех активных пользователей
	users, err := store.GetAllUsers()
	if err != nil {
		log.Fatalf("failed to get users: %v", err)
	}

	// Запуск горутин для каждого пользователя
	for _, chatID := range users {
		stopCh := make(chan bool)
		bot.StopChans[chatID] = stopCh

		go telegram.StartUserVacancyChecker(chatID, hhClient, store, bot, stopCh)
	}

	// Блокировка main потока
	select {}
}
