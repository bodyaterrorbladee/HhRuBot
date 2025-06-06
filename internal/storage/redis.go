package storage

import (
	"context"
	"fmt"
	"log"
	"strconv"
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
