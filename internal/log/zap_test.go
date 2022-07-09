package log_test

import (
	"go.uber.org/zap"
	"testing"

	"github.com/alexadhy/shortener/internal/log"
)

func BenchmarkLogger(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	log.WithFields(zap.String("app", "test"))
	for i := 0; i < b.N; i++ {
		log.Infof("got %d", i)
	}
}
