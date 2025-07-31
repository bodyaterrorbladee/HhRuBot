package main

import (

	"log"

	"hhruBot/internal/config"
	"hhruBot/internal/hh"
	"hhruBot/internal/storage"
	"hhruBot/internal/telegram"
)
func main() {
	cfg := config.LoadConfig()
	storage := storage.NewStorage(cfg)
	bot := telegram.NewBot(cfg, storage)
	hhClient := hh.NewClient()

	go bot.Start()

	users, err := storage.GetAllUsers()
	if err != nil {
		log.Fatalf("Не удалось получить пользователей: %v", err)
	}

	for _, chatID := range users {
		bot.StopChans[chatID] = make(chan bool)
		go telegram.StartUserVacancyChecker(chatID, hhClient, storage, bot, bot.StopChans[chatID])
	}

// Чтобы main не завершился
select {}
}


