package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Locker interface {
	Acquire(ctx context.Context, resource string, ttl time.Duration) (bool, error)
	Release(ctx context.Context, resource string) error
}

type DistributedLock struct {
	client  *Client
	ownerID string
}

func NewDistributedLock(client *Client, ownerID string) *DistributedLock {
	return &DistributedLock{
		client:  client,
		ownerID: ownerID,
	}
}

func (l *DistributedLock) key(resource string) string {
	return fmt.Sprintf("%slock:%s", l.client.prefix, resource)
}

func (l *DistributedLock) Acquire(ctx context.Context, resource string, ttl time.Duration) (bool, error) {
	key := l.key(resource)
	ok, err := l.client.rdb.SetNX(ctx, key, l.ownerID, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock for %s: %w", resource, err)
	}
	return ok, nil
}

var releaseScript = redis.NewScript(`
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
`)

func (l *DistributedLock) Release(ctx context.Context, resource string) error {
	key := l.key(resource)
	_, err := releaseScript.Run(ctx, l.client.rdb, []string{key}, l.ownerID).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock for %s: %w", resource, err)
	}
	return nil
}

type NoopLocker struct{}

func (n *NoopLocker) Acquire(ctx context.Context, resource string, ttl time.Duration) (bool, error) {
	return true, nil
}

func (n *NoopLocker) Release(ctx context.Context, resource string) error {
	return nil
}
