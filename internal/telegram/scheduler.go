package telegram

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"hhruBot/internal/hh"
	"hhruBot/internal/storage"
)

func StartUserVacancyChecker(
	chatID int64,
	hhClient *hh.Client,
	storage *storage.Storage,
	bot *Bot,
	stopCh chan bool,
) {
	intervalMin, err := storage.GetUserInterval(chatID)
	if err != nil || intervalMin <= 0 {
		intervalMin = 30
	}

	ticker := time.NewTicker(time.Duration(intervalMin) * time.Minute)
	defer ticker.Stop()

	lastChecked, err := storage.GetLastChecked(chatID)
	if err != nil {
		lastChecked = time.Now().Add(-time.Duration(intervalMin) * time.Minute)
	}

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			now := time.Now()

			if err := checkVacancies(chatID, hhClient, storage, bot, lastChecked); err != nil {
				log.Printf("❌ Ошибка при проверке вакансий [%d]: %v", chatID, err)
			} else {
				storage.SetLastChecked(chatID, now)
				lastChecked = now
			}
		}
	}
}

func checkVacancies(
	chatID int64,
	hhClient *hh.Client,
	storage *storage.Storage,
	bot *Bot,
	from time.Time,
) error {
	tagsStr, _ := storage.GetUserSetting(chatID, "tags")
	citiesStr, _ := storage.GetUserSetting(chatID, "cities")

	tags := parseCSV(tagsStr)
	cities := parseCSV(citiesStr)

	vacancies, err := hhClient.GetVacancies(tags, cities, from)
	if err != nil {
		return err
	}

	if len(vacancies) == 0 {
		bot.SendMessage(chatID, "🔍 Новые вакансии не были найдены.")
		return nil
	}

	for _, v := range vacancies {
		vacID, err := strconv.Atoi(v.Id)
		if err != nil || storage.AlreadySeen(vacID) {
			continue
		}

		url := fmt.Sprintf("https://hh.ru/vacancy/%s", v.Id)
		text := fmt.Sprintf("❤ *Новая вакансия:* [%s](%s)\n🏙️ Город: %s", v.Name, url, v.Area.Name)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"

		if _, err := bot.Api.Send(msg); err != nil {
			log.Printf("❌ Не удалось отправить вакансию: %v", err)
			continue
		}

		storage.MarkAsSeen(vacID)
	}

	return nil
}

func parseCSV(input string) []string {
	var result []string
	for _, s := range strings.Split(input, ",") {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
