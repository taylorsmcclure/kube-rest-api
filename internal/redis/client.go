package k8sredis

import (
	"context"
	"crypto/tls"

	"github.com/go-redis/redis/v8"
	e "github.com/taylorsmcclure/kube-server/internal/errors"
	"github.com/taylorsmcclure/kube-server/internal/logger"
)

// Creates a Redis client with mTLS authentication
func NewClient(address string, redisTLSConfig *tls.Config) (*redis.Client, error) {
	defer e.NonFatal()

	ctx := context.Background()

	rClient := redis.NewClient(&redis.Options{
		Addr:      address,
		TLSConfig: redisTLSConfig,
		DB:        0, // use default DB
	})

	// Verify we can connect to Redis
	_, err := rClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	logger.Log.Infof("Authenticated to Redis at : %s", address)

	return rClient, nil
}
