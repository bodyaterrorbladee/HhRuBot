package storage

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"
	"strings"

	"github.com/redis/go-redis/v9"
	"hhruBot/internal/config"
)

type Storage struct {
	client *redis.Client
	ctx    context.Context
}

func NewStorage(cfg *config.Config) *Storage {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Не удалось подключиться к Redis:", err)
	}

	fmt.Println("✅ Успешное подключение к Redis")

	return &Storage{client: rdb, ctx: ctx}
}

// === Вакансии ===

func (s *Storage) AlreadySeen(vacancyID int) bool {
	key := "vacancy:" + strconv.Itoa(vacancyID)
	exists, err := s.client.Exists(s.ctx, key).Result()
	if err != nil {
		log.Printf("Redis Exists error: %v", err)
		return false
	}
	return exists == 1
}

func (s *Storage) MarkAsSeen(vacancyID int) {
	key := "vacancy:" + strconv.Itoa(vacancyID)
	if err := s.client.Set(s.ctx, key, "1", 7*24*time.Hour).Err(); err != nil {
		log.Printf("Redis Set error: %v", err)
	}
}

// === Настройки пользователя ===

func (s *Storage) SetUserSetting(chatID int64, key, value string) error {
	redisKey := fmt.Sprintf("user:%d:%s", chatID, key)
	return s.client.Set(s.ctx, redisKey, value, 0).Err()
}

func (s *Storage) GetUserSetting(chatID int64, key string) (string, error) {
	redisKey := fmt.Sprintf("user:%d:%s", chatID, key)
	return s.client.Get(s.ctx, redisKey).Result()
}

// Удобный метод для интервала как int
func (s *Storage) GetUserInterval(chatID int64) (int, error) {
	val, err := s.GetUserSetting(chatID, "interval")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

// === Последнее время проверки ===

func (s *Storage) SetLastChecked(chatID int64, t time.Time) error {
	key := fmt.Sprintf("user:%d:last_checked", chatID)
	return s.client.Set(s.ctx, key, t.Unix(), 0).Err()
}

func (s *Storage) GetLastChecked(chatID int64) (time.Time, error) {
	key := fmt.Sprintf("user:%d:last_checked", chatID)
	unixTs, err := s.client.Get(s.ctx, key).Int64()
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(unixTs, 0), nil
}

// === Пользователи ===

func (s *Storage) AddUser(chatID int64) error {
	return s.client.SAdd(s.ctx, "users", strconv.FormatInt(chatID, 10)).Err()
}

func (s *Storage) GetUsers() ([]int64, error) {
	userStrs, err := s.client.SMembers(s.ctx, "users").Result()
	if err != nil {
		return nil, err
	}

	users := make([]int64, 0, len(userStrs))
	for _, u := range userStrs {
		id, err := strconv.ParseInt(u, 10, 64)
		if err == nil {
			users = append(users, id)
		}
	}
	return users, nil
}
func (s *Storage) PauseUser(chatID int64) error {
	key := fmt.Sprintf("user:%d:paused", chatID)
	return s.client.Set(s.ctx, key, "1", 0).Err()
}

func (s *Storage) ResumeUser(chatID int64) error {
	key := fmt.Sprintf("user:%d:paused", chatID)
	return s.client.Del(s.ctx, key).Err()
}

func (s *Storage) IsUserPaused(chatID int64) (bool, error) {
	key := fmt.Sprintf("user:%d:paused", chatID)
	exists, err := s.client.Exists(s.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}
// Пометить пользователя как disabled (например, заблокировал бота).
func (s *Storage) DisableUser(chatID int64) error {
	key := fmt.Sprintf("user:%d:disabled", chatID)
	if err := s.client.Set(s.ctx, key, "1", 0).Err(); err != nil {
		return err
	}
	// опционально удаляем из множества активных пользователей
	return s.client.SRem(s.ctx, "users", strconv.FormatInt(chatID, 10)).Err()
}

func (s *Storage) IsUserDisabled(chatID int64) (bool, error) {
	key := fmt.Sprintf("user:%d:disabled", chatID)
	exists, err := s.client.Exists(s.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists == 1, nil
}
func (s *Storage) GetActiveUsers() ([]int64, error) {
    keys, err := s.client.Keys(s.ctx, "user:*:paused").Result()
    if err != nil {
        return nil, err
    }

    var activeUsers []int64
    for _, key := range keys {
        paused, err := s.client.Get(s.ctx, key).Result()
        if err != nil && err != redis.Nil {
            continue
        }
        if paused == "1" {
            continue
        }

        // key example: user:12345:paused → extract 12345
        parts := strings.Split(key, ":")
        if len(parts) < 2 {
            continue
        }
        chatID, err := strconv.ParseInt(parts[1], 10, 64)
        if err != nil {
            continue
        }
        activeUsers = append(activeUsers, chatID)
    }

    return activeUsers, nil
}
