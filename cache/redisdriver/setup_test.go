package redisdriver

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testRedisCache RedisCache

func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start Redis container
	redisContainer, err := redis.Run(ctx,
		"redis:7-alpine",
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("Failed to start Redis container: %v", err)
	}
	defer redisContainer.Terminate(ctx)

	// Get connection details
	host, err := redisContainer.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get host: %v", err)
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		log.Fatalf("Failed to get port: %v", err)
	}

	pool, err := CreateRedisPool("10", "100", "240",
		host+":"+port.Port(), "", "")
	if err != nil {
		log.Fatalf("Failed to create Redis pool: %v", err)
	}
	defer pool.Close()

	// Create Redis cache
	testRedisCache = RedisCache{
		Conn:   pool,
		Prefix: "test",
	}

	os.Exit(m.Run())
}
