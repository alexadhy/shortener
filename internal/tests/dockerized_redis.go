package tests

import (
	"errors"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/alexadhy/shortener/persist/redis"
)

const (
	expirationInSeconds = 120
	poolMaxWait         = 120 * time.Second
)

func BootstrapRedis() (*redis.Store, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, err
	}

	// pulls an image, creates a container based on it and runs
	res, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "7-alpine",
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, err
	}

	if err := res.Expire(expirationInSeconds); err != nil {
		return nil, err
	}

	var store *redis.Store

	pool.MaxWait = poolMaxWait
	if err = pool.Retry(func() error {
		time.Sleep(4 * time.Second)
		addr := res.GetHostPort("6379/tcp")
		store, err = redis.New(addr)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if store == nil {
		return nil, errors.New("cannot connect to redis")
	}

	return store, nil
}
