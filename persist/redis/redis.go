package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/alexadhy/shortener/model"
)

// Store implements persist.Persist
type Store struct {
	rc redis.UniversalClient
}

// Get the value of a shortened url from redis
func (s *Store) Get(ctx context.Context, key string) (*model.ShortenedData, error) {
	val, err := s.rc.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var m model.ShortenedData
	if _, err = m.UnmarshalMsg([]byte(val)); err != nil {
		return nil, err
	}
	m.Key = key
	m.Expiry = m.Expiry.UTC()
	return &m, nil
}

// Set the value of a shortened url to redis, while checking for duplicates
func (s *Store) Set(ctx context.Context, data *model.ShortenedData) error {
	_, err := s.Get(ctx, data.Key)
	if err == redis.Nil {
		exp := data.Expiry.Sub(time.Now().UTC())
		b, err := data.MarshalMsg(nil)
		if err != nil {
			return err
		}
		if err = s.rc.SetEX(ctx, data.Key, b, exp).Err(); err != nil {
			return err
		}
	}
	return nil
}

// Expire will not do anything on redis since we set the expiry from redis.SetEX
func (s *Store) Expire(_ context.Context) (int, error) {
	return 0, nil
}

// New creates a new instance of *Store
func New(addresses ...string) (*Store, error) {
	rc := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs: addresses,
		DB:    0,
	})

	if err := rc.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &Store{rc}, nil
}
