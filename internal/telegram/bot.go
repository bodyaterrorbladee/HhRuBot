package telegram

import (

	"log"
	"strconv"
	"strings"

	"hhruBot/internal/config"
	"hhruBot/internal/storage"
	"hhruBot/internal/hh"
	
	

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


type Bot struct {
	Api     *tgbotapi.BotAPI
	Storage *storage.Storage
	StopChans map[int64]chan bool
}


func NewBot(cfg *config.Config, storage *storage.Storage) *Bot {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatal("Не удалось создать бота")
	}

	log.Printf("Авторизация прошла как: %s", api.Self.UserName)

	return &Bot{
		Api:       api,
		Storage:   storage,
		StopChans: make(map[int64]chan bool),
	}
}

func (b *Bot) SendMessage(chatID int64, text string){
	msg := tgbotapi.NewMessage(chatID, text)
	_,err := b.Api.Send(msg)
	if err!=nil{
		log.Print("Не удалось отправить сообщение")
	}
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.Api.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		text := update.Message.Text

		switch {
		case strings.HasPrefix(text, "/start"):
			b.SendMessage(chatID, "Привет! Я бот для поиска вакансий. Используй команды:\n"+
				"/tags golang,devops\n"+
				"/city Москва,Санкт-Петербург\n"+
				"/interval 30")

		case strings.HasPrefix(text, "/tags"):
			tags := strings.TrimSpace(strings.TrimPrefix(text, "/tags"))
			if tags == "" {
				b.SendMessage(chatID, "Введите ключевые слова после команды, пример:\n/tags golang,devops")
				continue
			}
			err := b.Storage.SetUserSetting(chatID, "tags", tags)
			if err != nil {
				b.SendMessage(chatID, "Ошибка при сохранении тегов")
			} else {
				_ = b.Storage.AddUser(chatID) // 👈 добавляем пользователя
				b.SendMessage(chatID, "Теги сохранены: "+tags)
}

		case strings.HasPrefix(text, "/city"):
			cities := strings.TrimSpace(strings.TrimPrefix(text, "/city"))
			if cities == "" {
				b.SendMessage(chatID, "Введите города после команды, пример:\n/city Москва,Санкт-Петербург")
				continue
			}
			err := b.Storage.SetUserSetting(chatID, "cities", cities)
				if err != nil {
					b.SendMessage(chatID, "Ошибка при сохранении городов")
				} else {
					_ = b.Storage.AddUser(chatID) // 👈 добавляем пользователя
					b.SendMessage(chatID, "Города сохранены: "+cities)
}

case strings.HasPrefix(text, "/interval"):
	intervalStr := strings.TrimSpace(strings.TrimPrefix(text, "/interval"))
	intervalMin, err := strconv.Atoi(intervalStr)
	if err != nil || intervalMin <= 0 {
		b.SendMessage(chatID, "Интервал должен быть положительным числом в минутах, пример:\n/interval 30")
		continue
	}

	err = b.Storage.SetUserSetting(chatID, "interval", intervalStr)
	if err != nil {
		b.SendMessage(chatID, "Ошибка при сохранении интервала")
		return
	}

	b.SendMessage(chatID, "Интервал сохранён: "+intervalStr+" мин.")

	// 🔁 Останавливаем старую горутину, если есть
	if stopCh, ok := b.StopChans[chatID]; ok {
		stopCh <- true
	}
	newStopCh := make(chan bool)
	b.StopChans[chatID] = newStopCh
	go StartUserVacancyChecker(chatID, hh.NewClient(), b.Storage, b, newStopCh)

		default:
			b.SendMessage(chatID, "Неизвестная команда. Попробуйте /start")
		}
	}
}
