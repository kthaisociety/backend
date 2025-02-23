package database

import (
	"fmt"
	"sync"

	"backend/internal/config"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

func GetRedisClient(cfg *config.Config) (*redis.Client, error) {
	var err error
	
	redisOnce.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		})
	})

	return redisClient, err
} 