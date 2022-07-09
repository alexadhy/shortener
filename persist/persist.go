package persist

import (
	"context"

	"github.com/alexadhy/shortener/model"
)

// Persist is the common interface to all of the storage type that interact with *model.ShortenedData
type Persist interface {
	// Get the value of a shortened url from the persistence layer
	Get(ctx context.Context, key string) (*model.ShortenedData, error)
	// Set the value of a shortened url to the persistence layer, while checking for duplicates
	Set(ctx context.Context, data *model.ShortenedData) error
	// Expire will evict the data of a shortened url from the persistence layer
	Expire(ctx context.Context) (int, error)
}
