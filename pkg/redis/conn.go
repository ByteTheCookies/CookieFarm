package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type redisConn struct {
	addr     string        // Redis server address
	password string        // Password for Redis server, if any
	db       int           // Redis database number, default is 0
	conn     *redis.Client // Redis client connection
}

func (r redisConn) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func GetDefaultRedisConn() *redisConn {
	return &redisConn{
		addr:     "localhost:6379", // Default address
		password: "",               // Default no password
		db:       0,                // Default database
		conn:     nil,              // No connection yet
	}
}

func GetRedisConn(addr, password string, db int) *redisConn {
	return &redisConn{
		addr:     addr,     // Default address
		password: password, // Default no password
		db:       db,       // Default database
	}
}

func (r redisConn) NewClient() (*redis.Client, error) {
	ctx := context.Background()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     r.addr,
		Password: r.password,
		DB:       r.db,
	})

	// Test the connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	r.conn = redisClient
	return redisClient, nil
}
