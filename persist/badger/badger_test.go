package badger_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
	"time"

	bd "github.com/dgraph-io/badger/v3"

	"github.com/alexadhy/shortener/model"
	"github.com/alexadhy/shortener/persist/badger"
)

func seedDataToDB(t *testing.T, n int, store *badger.Store) []*model.ShortenedData {
	fakeData, err := model.GenFake(n)
	if err != nil {
		t.Fatalf("seedDataToDB(): %v", err)
	}
	for _, f := range fakeData {
		if err = store.Set(context.Background(), f); err != nil {
			t.Fatalf("seedDataToDB() Set: %v", err)
		}
	}
	return fakeData
}

func bootstrapBadger(t *testing.T) *badger.Store {
	wd, _ := os.Getwd()
	fpath := filepath.Join(wd, "badger-test")

	s, err := badger.New(fpath)
	if err != nil {
		t.Fatalf("error initiating store: %v", err)
	}
	return s
}

func TestGet(t *testing.T) {
	s := bootstrapBadger(t)
	defer s.Shutdown()
	fakeDatas := seedDataToDB(t, 3, s)

	cases := []struct {
		name      string
		input     string
		want      *model.ShortenedData
		wantError error
	}{
		{
			name:      "should be able to correctly get data if key is valid",
			input:     fakeDatas[0].Key,
			want:      fakeDatas[0],
			wantError: nil,
		},
		{
			name:      "should return error if key doesn't exist",
			input:     "aBCV3441",
			want:      nil,
			wantError: bd.ErrKeyNotFound,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.Get(context.TODO(), tt.input)
			if tt.wantError != nil {
				assert.Equal(t, err, tt.wantError)
			}

			if tt.want != nil {
				assert.Equal(t, *tt.want, *got)
			}
		})
	}
}

func TestSet(t *testing.T) {
	s := bootstrapBadger(t)
	defer s.Shutdown()
	existingData := seedDataToDB(t, 1, s)

	fakeDatas, err := model.GenFake(3)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name   string
		input  *model.ShortenedData
		hasErr bool
	}{
		{
			name:   "should be able to correctly set data if data is valid",
			input:  fakeDatas[0],
			hasErr: false,
		},
		{
			name: "should not be able to set data if expiry is wrong",
			input: &model.ShortenedData{
				Key:    fakeDatas[1].Key,
				Hash:   fakeDatas[1].Hash,
				Short:  fakeDatas[1].Short,
				Orig:   fakeDatas[1].Orig,
				Expiry: time.Now().AddDate(0, 0, -1).UTC(),
			},
			hasErr: true,
		},
		{
			name:   "trying to input the same data twice doesn't result in any error",
			input:  existingData[0],
			hasErr: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err = s.Set(context.Background(), tt.input)
			if tt.hasErr {
				assert.NotNil(t, err)
				t.Log(err)
			}

		})
	}
}
