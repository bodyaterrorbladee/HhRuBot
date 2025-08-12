package telegram

import (
	"log"
	"strconv"
	"strings"

	"hhruBot/internal/config"
	"hhruBot/internal/hh"
	"hhruBot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	Api       *tgbotapi.BotAPI
	Storage   *storage.Storage
	HHClient  *hh.Client
	StopChans map[int64]chan struct{}
}

func NewBot(cfg *config.Config, storage *storage.Storage) *Bot {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Fatal("Не удалось создать бота")
	}

	log.Printf("Авторизация прошла как: %s", api.Self.UserName)

	b := &Bot{
		Api:       api,
		Storage:   storage,
		HHClient:  hh.NewClient(),
		StopChans: make(map[int64]chan struct{}),
	}

	// 🔄 Автоматический запуск чекеров для активных пользователей
	users, err := storage.GetActiveUsers()
	if err != nil {
		log.Printf("⚠ Не удалось получить список активных пользователей: %v", err)
	} else {
		for _, chatID := range users {
			log.Printf("▶ Автозапуск чекера для chatID %d", chatID)
			stopCh := make(chan struct{})
			b.StopChans[chatID] = stopCh
			go StartUserVacancyChecker(chatID, b.HHClient, b.Storage, b, stopCh)
		}
	}

	return b
}

func (b *Bot) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.Api.Send(msg)
	if err != nil {
		log.Printf("Не удалось отправить сообщение %d: %v", chatID, err)
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
			b.Storage.AddUser(chatID)
			b.SendMessage(chatID, `👋 Добро пожаловать в HH.ru Бот!

Я помогу тебе следить за новыми вакансиями.

⚙️ Основные команды:
/tags golang,devops — задать ключевые слова
/city Москва — выбрать город(а)
/interval 30 — интервал проверки (в минутах)
/pause — приостановить уведомления
/search — возобновить работу
/settings — показать текущие настройки
/help — справка по командам`)

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
				_ = b.Storage.AddUser(chatID)
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
				_ = b.Storage.AddUser(chatID)
				b.SendMessage(chatID, "Города сохранены: "+cities)
			}

		case strings.HasPrefix(text, "/interval"):
			intervalStr := strings.TrimSpace(strings.TrimPrefix(text, "/interval"))
			intervalMin, err := strconv.Atoi(intervalStr)
			if err != nil || intervalMin <= 0 {
				b.SendMessage(chatID, "Интервал должен быть положительным числом в минутах, пример:\n/interval 30")
				continue
			}

			if intervalMin < 5 {
				b.SendMessage(chatID, "⚠ Минимальный интервал — 5 минут.")
				continue
			}

			err = b.Storage.SetUserSetting(chatID, "interval", intervalStr)
			if err != nil {
				b.SendMessage(chatID, "Ошибка при сохранении интервала")
				return
			}

			b.SendMessage(chatID, "Интервал сохранён: "+intervalStr+" мин.")

			if stopCh, ok := b.StopChans[chatID]; ok {
				close(stopCh)
			}
			newStopCh := make(chan struct{})
			b.StopChans[chatID] = newStopCh
			go StartUserVacancyChecker(chatID, b.HHClient, b.Storage, b, newStopCh)

		case strings.HasPrefix(text, "/settings"):
			tags, _ := b.Storage.GetUserSetting(chatID, "tags")
			cities, _ := b.Storage.GetUserSetting(chatID, "cities")
			interval, _ := b.Storage.GetUserSetting(chatID, "interval")

			if tags == "" {
				tags = "не установлены"
			}
			if cities == "" {
				cities = "не установлены"
			}

			if interval == "" {
				interval = "по умолчанию (30 минут)"
			} else {
				if val, err := strconv.Atoi(interval); err == nil {
					if val < 5 {
						interval = "по умолчанию (5 минут)"
					} else {
						interval += " минут"
					}
				} else {
					interval = "по умолчанию (30 минут)"
				}
			}

			settingsMsg := "📌 *Ваши настройки:*\n" +
				"🔖 Теги: `" + tags + "`\n" +
				"🏙️ Города: `" + cities + "`\n" +
				"⏱️ Интервал: `" + interval + "`"

			msg := tgbotapi.NewMessage(chatID, settingsMsg)
			msg.ParseMode = "Markdown"
			_, _ = b.Api.Send(msg)

		case strings.HasPrefix(text, "/pause"):
			err := b.Storage.PauseUser(chatID)
			if err != nil {
				b.SendMessage(chatID, "❌ Не удалось поставить на паузу.")
				continue
			}
			if stopCh, ok := b.StopChans[chatID]; ok {
				close(stopCh)
				delete(b.StopChans, chatID)
			}
			b.SendMessage(chatID, "⏸️ Поиск вакансий приостановлен. Для продолжения — /search.")

		case strings.HasPrefix(text, "/search"):
			paused, _ := b.Storage.IsUserPaused(chatID)
			if !paused {
				b.SendMessage(chatID, "🔄 Поиск уже активен.")
				continue
			}

			err := b.Storage.ResumeUser(chatID)
			if err != nil {
				b.SendMessage(chatID, "❌ Не удалось возобновить поиск.")
				continue
			}

			stopCh := make(chan struct{})
			b.StopChans[chatID] = stopCh
			go StartUserVacancyChecker(chatID, b.HHClient, b.Storage, b, stopCh)

			b.SendMessage(chatID, "✅ Поиск возобновлён.")

		case strings.HasPrefix(text, "/help"):
			b.SendMessage(chatID, `🛠 Доступные команды:
/tags — задать ключевые слова
/city — выбрать города
/interval — частота поиска (в минутах)
/pause — остановить рассылку
/search — возобновить рассылку
/settings — показать текущие настройки
/help — показать справку`)

		default:
			b.SendMessage(chatID, "Неизвестная команда. Попробуйте /start")
		}
	}
}
