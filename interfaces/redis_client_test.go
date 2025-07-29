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

	err := redisClient.Set(context.Background(), "test-key", "test-value", time.Hour)
	assert.NoError(t, err)

	value, err := redisClient.Get(context.Background(), "test-key")
	assert.NoError(t, err)
	assert.Equal(t, "test-value", value)

	err = redisClient.Del(context.Background(), "test-key")
	assert.NoError(t, err)
}
