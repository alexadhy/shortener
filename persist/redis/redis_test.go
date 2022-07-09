package redis_test

//lint:file-ignore SA3001 i want to set the amount of operations myself

import (
	"context"
	"reflect"
	"testing"
	"time"

	rd "github.com/go-redis/redis/v8"

	"github.com/alexadhy/shortener/internal/tests"
	"github.com/alexadhy/shortener/model"
	"github.com/alexadhy/shortener/persist/redis"
)

func seedDataToDB(t *testing.T, n int, store *redis.Store) []*model.ShortenedData {
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

func TestGet(t *testing.T) {
	s, err := tests.BootstrapRedis()
	if err != nil {
		t.Fatal(err)
	}
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
			wantError: rd.Nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.Get(context.Background(), tt.input)
			if tt.wantError != nil {
				if !reflect.DeepEqual(err, tt.wantError) {
					t.Fatalf("expecting error to be %v, got: %v", tt.wantError, err)
				}
			}
			if tt.want != nil {
				if err != nil {
					t.Fatalf("expecting no error, instead got: %v", err)
				}
				if !reflect.DeepEqual(*tt.want, *got) {
					t.Fatalf("expecting %v\n got: %v", *tt.want, *got)
				}
			}
		})
	}
}

func TestSet(t *testing.T) {
	s, err := tests.BootstrapRedis()
	if err != nil {
		t.Fatal(err)
	}
	existingData := seedDataToDB(t, 1, s)

	fakeDatas, err := model.GenFake(3)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name      string
		input     *model.ShortenedData
		wantError string
	}{
		{
			name:      "should be able to correctly set data if data is valid",
			input:     fakeDatas[0],
			wantError: "",
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
			wantError: "ERR invalid expire time in 'setex' command",
		},
		{
			name:      "trying to input the same data twice doesn't result in any error",
			input:     existingData[0],
			wantError: "",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := s.Set(context.Background(), tt.input)
			if tt.wantError != "" {
				if !reflect.DeepEqual(err.Error(), tt.wantError) {
					t.Fatalf("expecting error to be %s, got: %s", tt.wantError, err)
				}
			}
		})
	}
}

func BenchmarkSet(b *testing.B) {
	s, err := tests.BootstrapRedis()
	if err != nil {
		b.Fatal(err)
	}
	fakedata, err := model.GenFake(100)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.N = len(fakedata)
	for i := 0; i < b.N; i++ {
		_ = s.Set(context.Background(), fakedata[i])
	}
}
