package middlewares

import (
	"net/http"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
)

// LimitHandler creates a new rate-limiter using tollbooth
func LimitHandler(lmt *limiter.Limiter) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		wrapper := &limiterWrapper{
			lmt: lmt,
		}

		wrapper.handler = handler
		return wrapper
	}
}

type limiterWrapper struct {
	lmt     *limiter.Limiter
	handler http.Handler
}

func (l *limiterWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	select {
	case <-ctx.Done():
		http.Error(w, "Context was canceled", http.StatusServiceUnavailable)
		return

	default:
		err := tollbooth.LimitByRequest(l.lmt, w, r)
		if err != nil {
			w.Header().Add("Content-Type", l.lmt.GetMessageContentType())
			w.WriteHeader(err.StatusCode)
			_, _ = w.Write([]byte(err.Message))
			return
		}

		l.handler.ServeHTTP(w, r)
	}
}
