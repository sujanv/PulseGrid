package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sujanv/PulseGrid/internal/config"
	"github.com/sujanv/PulseGrid/internal/models"
)

type Client struct {
	rdb    *redis.Client
	prefix string
}

func NewClient(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Client{
		rdb:    rdb,
		prefix: cfg.KeyPrefix,
	}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Client) SetTargetState(ctx context.Context, targetID string, status models.CheckStatus) error {
	key := fmt.Sprintf("%starget:%s:state", c.prefix, targetID)
	return c.rdb.Set(ctx, key, string(status), 24*time.Hour).Err()
}

func (c *Client) GetTargetState(ctx context.Context, targetID string) (models.CheckStatus, error) {
	key := fmt.Sprintf("%starget:%s:state", c.prefix, targetID)
	val, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return models.CheckStatus(val), nil
}

func (c *Client) IncrConsecutiveFailures(ctx context.Context, targetID string) (int64, error) {
	key := fmt.Sprintf("%starget:%s:fails", c.prefix, targetID)
	val, err := c.rdb.Incr(ctx, key).Result()
	if err == nil {
		c.rdb.Expire(ctx, key, 24*time.Hour)
	}
	return val, err
}

func (c *Client) ResetConsecutiveFailures(ctx context.Context, targetID string) error {
	key := fmt.Sprintf("%starget:%s:fails", c.prefix, targetID)
	return c.rdb.Del(ctx, key).Err()
}

func (c *Client) GetConsecutiveFailures(ctx context.Context, targetID string) (int64, error) {
	key := fmt.Sprintf("%starget:%s:fails", c.prefix, targetID)
	val, err := c.rdb.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (c *Client) RawClient() *redis.Client {
	return c.rdb
}
