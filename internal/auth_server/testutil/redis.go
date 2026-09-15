package testutil

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"

	"github.com/Nurlan270/cloud-storage-go/internal/auth_server/testutil/closer"
	coreredis "github.com/Nurlan270/cloud-storage-go/internal/core/redis"

	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func NewTestRedis(ctx context.Context, conf coreredis.Config) (*redis.Client, error) {
	rc, err := tcredis.Run(ctx,
		"redis:8.4-alpine",
		tcredis.WithSnapshotting(10, 1),
	)
	if err != nil {
		return nil, fmt.Errorf("redis: failed to start container: %s", err)
	}

	//	Get Host/Port pair from container
	redisHostPort, err := rc.Endpoint(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve redis host/port pair: %s", err)
	}

	redisHost, redisPort, err := net.SplitHostPort(redisHostPort)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis host and port: %s", err)
	}

	conf.Host = redisHost
	redisPortUInt, _ := strconv.ParseUint(redisPort, 10, 16)
	conf.Port = uint16(redisPortUInt)

	rdb, err := coreredis.Connect(conf, 0)
	if err != nil {
		return nil, err
	}

	closer.Add("Test Redis", func() error {
		return errors.Join(
			rdb.Close(),
			terminateTestRedis(rc),
		)
	})

	return rdb, nil
}

func CleanupRedis(t *testing.T, rdb *redis.Client) {
	t.Helper()

	if err := rdb.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("redis: failed to flush database: %s", err)
	}
}

func terminateTestRedis(rc *tcredis.RedisContainer) error {
	if err := testcontainers.TerminateContainer(rc); err != nil {
		return fmt.Errorf("redis: failed to terminate container: %s", err)
	}

	return nil
}
