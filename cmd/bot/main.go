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

	// Инициализация карт городов
	if err := hh.InitCityMap(); err != nil {
		log.Fatalf("Не удалось инициализировать карту городов: %v", err)
	}
	// Инициализация клиента HH.ru
	hhClient := hh.NewClient()

	// Запуск бота в отдельной горутине
	go bot.Start()

	// Получение всех активных пользователей
	users, err := store.GetUsers()
	if err != nil {
		log.Fatalf("failed to get users: %v", err)
	}

	// Запуск горутин для каждого пользователя
	for _, chatID := range users {
		bot.StopChans[chatID] = make(chan struct{})
go telegram.StartUserVacancyChecker(chatID, hhClient, store, bot, bot.StopChans[chatID])
	}

	// Блокировка main потока
	select {}
}
