package telegram

import (
	"log"
	"errors"
	"strconv"

	"hhruBot/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)


type Bot struct{
	Api *tgbotapi.BotAPI
	ChatID int64
}
func parseChatID(s string) (int64,error){
	id, err := strconv.ParseInt(s,10,64)
	if err!=nil{
		return 0, errors.New("chat_id должен быть числом")
	}
	return id, nil
}

func NewBot(cfg *config.Config) *Bot{
	api,err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err!=nil{
		log.Fatal("Не удалось создать бота")

	}
	chatID,err := parseChatID(cfg.TelegramChatId)
	if err!=nil{
		log.Fatal("Неверный chat_id")
	}

	return &Bot{
		Api: api,
		ChatID: chatID,
	}
}

func (b *Bot) SendMessage(text string){
	msg := tgbotapi.NewMessage(b.ChatID,text)
	_,err := b.Api.Send(msg)
	if err!=nil{
		log.Print("Не удалось отправить сообщение")
	}
}

