package telegram

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"hhruBot/internal/hh"
	"hhruBot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartUserVacancyChecker(chatID int64, hhClient *hh.Client, storage *storage.Storage, bot *Bot, stopCh chan bool) {
	intervalStr, _ := storage.GetUserSetting(chatID, "interval")
	intervalMin, err := strconv.Atoi(intervalStr)
	if err != nil || intervalMin <= 0 {
		intervalMin = 30 // по умолчанию
	}

	ticker := time.NewTicker(time.Duration(intervalMin) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		default:
			checkVacanciesForUser(chatID, hhClient, storage, bot)
			time.Sleep(time.Duration(intervalMin) * time.Minute)
		}
	}
}

func checkVacanciesForUser(chatID int64, hhClient *hh.Client, storage *storage.Storage, bot *Bot) {
	tags, _ := storage.GetUserSetting(chatID, "tags")
	cities, _ := storage.GetUserSetting(chatID, "cities")

	vacancies, err := hhClient.GetVacancies(parseCSV(tags), parseCSV(cities))
	if err != nil {
		log.Println("Ошибка при получении вакансий:", err)
		return
	}

	for _, v := range vacancies {
		if !matchesFilter(v.Name, tags) || !matchesCity(v.Area.Name, cities) {
			continue
		}

		vacID, err := strconv.Atoi(v.Id)
		if err != nil {
			continue
		}
		if storage.AlreadySeen(vacID) {
			continue
		}

		url := fmt.Sprintf("https://hh.ru/vacancy/%s", v.Id)
		text := fmt.Sprintf("❤ *Новая вакансия:* [%s](%s)\n🏙️ Город: %s", v.Name, url, v.Area.Name)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"

		if _, err := bot.Api.Send(msg); err != nil {
			log.Println("Ошибка отправки:", err)
		}
		storage.MarkAsSeen(vacID)
	}
}

func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func matchesFilter(text, tagsCSV string) bool {
	if tagsCSV == "" {
		return true
	}
	tags := strings.Split(tagsCSV, ",")
	textLower := strings.ToLower(text)
	for _, tag := range tags {
		if strings.Contains(textLower, strings.ToLower(strings.TrimSpace(tag))) {
			return true
		}
	}
	return false
}

func matchesCity(cityName, citiesCSV string) bool {
	if citiesCSV == "" {
		return true
	}
	cities := strings.Split(citiesCSV, ",")
	for _, city := range cities {
		if strings.EqualFold(strings.TrimSpace(city), cityName) {
			return true
		}
	}
	return false
}
