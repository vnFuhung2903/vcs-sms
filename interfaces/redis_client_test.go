package interfaces

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisClient(t *testing.T) {
	opt := &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	rds := redis.NewClient(opt)
	assert.NotNil(t, rds)

	redisClient := NewRedisClient(rds)
	err := redisClient.Set(context.Background(), "test_key", "test_value", time.Hour)
	assert.Error(t, err)
	value, err := redisClient.Get(context.Background(), "test_key")
	assert.Error(t, err)
	assert.Equal(t, "", value)
}
