package storage

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

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
		log.Fatal("Ну удалось подключиться к Redis", err)
	}
	fmt.Println("Успешное подключение к Redis")

	return &Storage{
		client: rdb,
		ctx:    ctx,
	}

}

func (s *Storage) AlreadySeen(vacancyID int) bool {
	key := "vacancy:" + strconv.Itoa(vacancyID)

	exists, err := s.client.Exists(s.ctx, key).Result()
	if err != nil {
		log.Printf("Redis ошибка метода Exists:%v", err)
		return false
	}
	return exists == 1
}

func (s *Storage) MarkAsSeen(vacancyID int) {
	key := "vacancy:" + strconv.Itoa(vacancyID)

	err := s.client.Set(s.ctx, key, "1", 7*24*time.Hour).Err()

	if err != nil {
		log.Printf("Redis ошибка метода SET :%v", err)
	}
}
func (s *Storage) SetUserSetting(chatID int64, key, value string) error {
	fullKey := fmt.Sprintf("user:%d:%s", chatID, key)
	return s.client.Set(s.ctx, fullKey, value, 0).Err()
}

// Получение пользовательской настройки
func (s *Storage) GetUserSetting(chatID int64, key string) (string, error) {
	fullKey := fmt.Sprintf("user:%d:%s", chatID, key)
	return s.client.Get(s.ctx, fullKey).Result()
}
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
func (s *Storage) GetAllUsers() ([]int64, error) {
	var result []int64

	iter := s.client.Scan(s.ctx, 0, "settings:*", 0).Iterator()
	for iter.Next(s.ctx) {
		key := iter.Val()
		parts := strings.Split(key, ":")
		if len(parts) != 3 {
			continue
		}

		chatIDStr := parts[1]
		chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
		if err != nil {
			continue
		}

		// Добавляем только уникальные chatID
		if !contains(result, chatID) {
			result = append(result, chatID)
		}
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func contains(slice []int64, item int64) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}