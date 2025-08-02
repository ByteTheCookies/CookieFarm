package redis

import (
	"context"

	"github.com/ByteTheCookies/CookieFarm/pkg/models"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

const streamKey = "flags_stream"

func (rdb *RedisConn) Add(ctx context.Context, data models.ClientData) (string, error) {
	args := &redis.XAddArgs{
		Stream: streamKey,
		MaxLen: 1000,
		Approx: true,
		Values: map[string]any{
			"data": data,
		},
	}
	id, err := rdb.conn.XAdd(ctx, args).Result()
	if err != nil {
		return "", err
	}
	return id, nil
}

func (rdb *RedisConn) Read(ctx context.Context, streamKey string, lastID string) ([]redis.XMessage, error) {
	args := &redis.XReadArgs{
		Streams: []string{streamKey, lastID},
		Block:   0,  // No blocking
		Count:   10, // Read up to 10 messages
	}
	messages, err := rdb.conn.XRead(ctx, args).Result()
	if err != nil {
		return nil, err
	}
	if len(messages) == 0 {
		return nil, nil // No new messages
	}

	return messages[0].Messages, nil
}
