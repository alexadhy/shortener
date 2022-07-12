package badger

import (
	"context"
	"errors"
	"github.com/alexadhy/shortener/internal/log"
	"os"
	"time"

	"github.com/dgraph-io/badger/v3"

	"github.com/alexadhy/shortener/model"
)

// Store implements persist.Persist
type Store struct {
	db   *badger.DB
	tiki time.Ticker
}

func (s Store) Get(_ context.Context, key string) (*model.ShortenedData, error) {
	var sd model.ShortenedData
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			if _, err = sd.UnmarshalMsg(val); err != nil {
				return err
			}
			sd.Key = key
			sd.Expiry = sd.Expiry.UTC()
			return nil
		})
	})

	return &sd, err
}

func (s Store) Set(_ context.Context, data *model.ShortenedData) error {
	err := s.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(data.Key))
		if err == nil {
			return nil
		}
		if err != nil && errors.Is(err, badger.ErrKeyNotFound) {
			exp := data.Expiry.Sub(time.Now().UTC())
			b, err := data.MarshalMsg(nil)
			if err != nil {
				return err
			}

			if time.Now().After(data.Expiry) {
				return errors.New("expiry is not valid")
			}

			newEntry := badger.NewEntry([]byte(data.Key), b).WithTTL(exp)
			return txn.SetEntry(newEntry)
		}
		return err
	})
	return err
}

func (s Store) Expire(_ context.Context) (int, error) {
	return 0, nil
}

// New takes a path to the new store and create a store
func New(pth string) (*Store, error) {
	_, err := os.Stat(pth)
	if err != nil {
		_ = os.MkdirAll(pth, 0700)
	}

	opt := badger.DefaultOptions(pth).
		WithBlockCacheSize(100 << 20).
		WithLogger(log.New())
	db, err := badger.Open(opt)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(5 * time.Minute)

	go func() {
		for range ticker.C {
		again:
			err := db.RunValueLogGC(0.7)
			if err == nil {
				goto again
			}
		}
	}()

	return &Store{db: db}, nil
}

func (s Store) Shutdown() error {
	s.tiki.Stop()
	return s.db.Close()
}
