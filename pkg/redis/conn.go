package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConn represents a Redis connection manager
type RedisConn struct {
	addr     string        // Redis server address
	password string        // Password for Redis server, if any
	db       int           // Redis database number, default is 0
	conn     *redis.Client // Redis client connection
	mu       sync.RWMutex  // Mutex for thread-safe operations
}

// RedisConfig holds configuration for Redis connection
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int           // Connection pool size
	MinIdleConns int           // Minimum idle connections
	MaxRetries   int           // Maximum number of retries
	DialTimeout  time.Duration // Timeout for establishing connection
	ReadTimeout  time.Duration // Timeout for reading data
	WriteTimeout time.Duration // Timeout for writing data
}

// DefaultRedisConfig returns a default Redis configuration
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// NewRedisConnWithConfig creates a new Redis connection with custom configuration
func NewRedisConnWithConfig(config *RedisConfig) *RedisConn {
	if config == nil {
		config = DefaultRedisConfig()
	}
	return &RedisConn{
		addr:     config.Addr,
		password: config.Password,
		db:       config.DB,
	}
}

// NewRedisConn creates a new Redis connection with basic parameters
func NewRedisConn(addr, password string, db int) *RedisConn {
	return &RedisConn{
		addr:     addr,
		password: password,
		db:       db,
	}
}

// Connect establishes a connection to Redis with proper configuration
func (r *RedisConn) Connect(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.conn != nil {
		return nil // Already connected
	}

	config := DefaultRedisConfig()
	config.Addr = r.addr
	config.Password = r.password
	config.DB = r.db

	redisClient := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// Test the connection with context
	if err := redisClient.Ping(ctx).Err(); err != nil {
		redisClient.Close()
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	r.conn = redisClient
	return nil
}

// ConnectWithConfig establishes a connection with custom configuration
func (r *RedisConn) ConnectWithConfig(ctx context.Context, config *RedisConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.conn != nil {
		return nil // Already connected
	}

	if config == nil {
		config = DefaultRedisConfig()
	}

	// Update connection parameters
	r.addr = config.Addr
	r.password = config.Password
	r.db = config.DB

	redisClient := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		redisClient.Close()
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	r.conn = redisClient
	return nil
}

// GetClient returns the Redis client, establishing connection if needed
func (r *RedisConn) GetClient(ctx context.Context) (*redis.Client, error) {
	r.mu.RLock()
	if r.conn != nil {
		client := r.conn
		r.mu.RUnlock()
		return client, nil
	}
	r.mu.RUnlock()

	// Need to establish connection
	if err := r.Connect(ctx); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conn, nil
}

// IsConnected checks if the connection is active
func (r *RedisConn) IsConnected(ctx context.Context) bool {
	r.mu.RLock()
	client := r.conn
	r.mu.RUnlock()

	if client == nil {
		return false
	}

	return client.Ping(ctx).Err() == nil
}

// Close closes the Redis connection
func (r *RedisConn) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.conn != nil {
		err := r.conn.Close()
		r.conn = nil
		return err
	}
	return nil
}

// Reconnect closes the current connection and establishes a new one
func (r *RedisConn) Reconnect(ctx context.Context) error {
	if err := r.Close(); err != nil {
		return fmt.Errorf("failed to close existing connection: %w", err)
	}
	return r.Connect(ctx)
}

// HealthCheck performs a comprehensive health check
func (r *RedisConn) HealthCheck(ctx context.Context) error {
	client, err := r.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client: %w", err)
	}

	// Test basic operations
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Test set/get operation
	testKey := "health_check_test"
	testValue := "ok"

	if err := client.Set(ctx, testKey, testValue, time.Minute).Err(); err != nil {
		return fmt.Errorf("set operation failed: %w", err)
	}

	val, err := client.Get(ctx, testKey).Result()
	if err != nil {
		return fmt.Errorf("get operation failed: %w", err)
	}

	if val != testValue {
		return fmt.Errorf("health check failed: expected %s, got %s", testValue, val)
	}

	// Clean up test key
	client.Del(ctx, testKey)

	return nil
}

// GetStats returns connection statistics
func (r *RedisConn) GetStats() *redis.PoolStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.conn != nil {
		return r.conn.PoolStats()
	}
	return nil
}
