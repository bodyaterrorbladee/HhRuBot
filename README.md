# HH.ru Bot 

Телеграм-бот на Go, который раз в 30 минут проверяет свежие вакансии на hh.ru (по ключевым словам: `golang`, `go developer`) 
и отправляет их в Telegram(нужен чат ID).

##  Пример работы

```
❤ Новая вакансия: [Go Developer](https://hh.ru/vacancy/1234567?query=golang)
🏙️ Город: Москва
```

## Используемые технологии

- Язык: Go 1.21.4
- Telegram Bot API (`go-telegram-bot-api`)
- hh.ru API
- Redis (для хранения уже отправленных вакансий)
- dotenv-конфигурация

## Как запустить?

### 1. Клонируй проект

```bash
git clone https://github.com/bodyaterrorbladee/hhru-telegram-go-bot.git
cd hhru-telegram-go-bot
```

### 2. Создай `.env` файл

```env
TELEGRAM_TOKEN=твой_токен
TELEGRAM_CHAT_ID=твой_chat_id
REDIS_ADDR=localhost:6379
```

> `TELEGRAM_CHAT_ID` можно получить через @userinfobot

### 3. Запусти Redis

```bash
redis-server
```

> Windows: через WSL или Docker

### 4. Запусти бота

```bash
go run main.go
```

##  Интервал

По умолчанию бот проверяет вакансии раз в 30 минут. Изменяется в `main.go`:

```go
ticker := time.NewTicker(30 * time.Minute)
```

##  Как это работает

- Получает вакансии с hh.ru
- Фильтрует по ключевым словам и городам (Москва, СПб)
- Проверяет Redis на дубли
- Отправляет сообщение в Telegram
