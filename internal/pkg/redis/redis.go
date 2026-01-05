package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Client wraps redis.Client with helpers
type Client struct {
	*redis.Client
}

// Connect establishes a connection to Redis
func Connect(addr, password string, db int) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		PoolTimeout:  4 * time.Second,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Info().
		Str("addr", addr).
		Int("db", db).
		Msg("Redis connected successfully")

	return &Client{Client: client}, nil
}

// Close closes the Redis connection
func (c *Client) Close() error {
	if c.Client != nil {
		log.Info().Msg("Closing Redis connection")
		return c.Client.Close()
	}
	return nil
}

// SetWithExpiry sets a key-value pair with TTL
func (c *Client) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Duration) error {
	return c.Set(ctx, key, value, expiry).Err()
}

// GetString gets a string value
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	val, err := c.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return val, err
}

// Exists checks if a key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IncrWithExpiry increments a counter and sets expiry if key is new
func (c *Client) IncrWithExpiry(ctx context.Context, key string, expiry time.Duration) (int64, error) {
	pipe := c.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiry)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}

	return incr.Val(), nil
}

// DeletePattern deletes all keys matching a pattern
func (c *Client) DeletePattern(ctx context.Context, pattern string) error {
	iter := c.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.Del(ctx, iter.Val()).Err(); err != nil {
			log.Error().Err(err).Str("key", iter.Val()).Msg("Failed to delete key")
		}
	}
	return iter.Err()
}
