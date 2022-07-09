package render

import (
	"encoding/json"
	"net/http"
)

// Response is just a generic structure over response
type Response[DataType any] struct {
	Headers    map[string]string `json:"-"`
	StatusCode int               `json:"-"`
	Data       DataType          `json:"data,omitempty" qs:"-"`
	Err        error             `json:"error,omitempty"`
}

func render[T any](resp Response[T], w http.ResponseWriter) (int, error) {
	for k, v := range resp.Headers {
		w.Header().Add(k, v)
	}
	w.WriteHeader(resp.StatusCode)
	b, _ := json.Marshal(&resp)
	return w.Write(b)
}

func Render(resp Response[any], w http.ResponseWriter) (int, error) {
	return render[any](resp, w)
}
