package middlewares

import (
	"github.com/alexadhy/shortener/internal/log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// LoggerMW returns a logger middleware for net/http, that implements the http.Handler interface.
func LoggerMW() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()
			defer func() {
				log.WithFields(
					zap.String("proto", r.Proto),
					zap.String("path", r.URL.Path),
					zap.String("reqId", middleware.GetReqID(r.Context())),
					zap.Int64("lat", time.Since(t1).Milliseconds()),
					zap.Int("status", ww.Status()),
					zap.Int("size", ww.BytesWritten()),
				)
				ref := ww.Header().Get("Referer")
				if ref == "" {
					ref = r.Header.Get("Referer")
				}
				if ref != "" {
					log.WithFields(zap.String("ref", ref))
				}
				ua := ww.Header().Get("User-Agent")
				if ua == "" {
					ua = r.Header.Get("User-Agent")
				}
				if ua != "" {
					log.WithFields(zap.String("ua", ua))
				}
			}()
			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
