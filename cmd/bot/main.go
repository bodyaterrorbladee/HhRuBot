package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"hhruBot/internal/config"
	"hhruBot/internal/hh"
	"hhruBot/internal/storage"
	"hhruBot/internal/telegram"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	cfg := config.LoadConfig()

	// Подключаем Redis
	storage := storage.NewStorage(cfg)

	// Подключаем Telegram бота
	bot := telegram.NewBot(cfg)

	// Создаём клиент HH
	hhClient := hh.NewClient()

	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	checkVacancies(hhClient, storage, bot)

	for range ticker.C {
		checkVacancies(hhClient, storage, bot)
	}
}

func checkVacancies(hhClient *hh.Client, storage *storage.Storage, bot *telegram.Bot) {
	fmt.Println("🔍 Проверяем новые вакансии...")

	vacancies, err := hhClient.GetVacancies()
	if err != nil {
		log.Println("Ошибка при получении вакансий:", err)
		return
	}

	for _, v := range vacancies {

		vacID, err := strconv.Atoi(v.Id)
		if err != nil {
			log.Printf("Невалидный ID вакансии: %v", v.Id)
			continue
		}

		if !storage.AlreadySeen(vacID) {
			vacancyURL := fmt.Sprintf("https://hh.ru/vacancy/%s?query=golang&hhtmFrom=vacancy_search_list", v.Id)
			text := fmt.Sprintf("❤ *Новая вакансия:* [%s](%s)\n🏙️ Город: %s", v.Name, vacancyURL, v.Area.Name)


			msg := tgbotapi.NewMessage(bot.ChatID, text)
			msg.ParseMode = "Markdown"

			_, err := bot.Api.Send(msg)
			if err != nil {
				log.Println("Ошибка при отправке в Telegram:", err)
				continue
			}

			storage.MarkAsSeen(vacID)
			fmt.Println("📩 Отправлено:", v.Name)
		}
	}

	fmt.Println("✅ Проверка завершена")
}
