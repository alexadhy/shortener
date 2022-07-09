package middlewares

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

type ExampleDTO struct {
	Name string `json:"name"`
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestRecoverer(t *testing.T) {
	cases := []struct {
		name        string
		handlerFunc func(http.ResponseWriter, *http.Request)
	}{
		{
			name:        "should panic & recover if the handler itself returns panic",
			handlerFunc: func(http.ResponseWriter, *http.Request) { panic("panic on handler level") },
		},
		{
			name: "should panic & recover if an interface{} is string and is converted to wrong type",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				b := &bytes.Buffer{}
				var a any

				defer r.Body.Close()
				_, _ = io.Copy(b, r.Body)
				a = b.String()

				c := a.(ExampleDTO)
				fmt.Print(c.Name)
			},
		},
		{
			name: "should panic & recover if an interface is nil and handler try to access it",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				var a any
				b := a.(ExampleDTO)
				fmt.Print(b)
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()

			oldRecovererErrorWriter := recovererErrorWriter
			defer func() { recovererErrorWriter = oldRecovererErrorWriter }()
			buf := &bytes.Buffer{}
			recovererErrorWriter = buf

			r.Use(Recoverer())
			r.HandleFunc("/", tt.handlerFunc)

			ts := httptest.NewServer(r)
			defer ts.Close()

			res, _ := testRequest(t, ts, http.MethodGet, "/", nil)
			assertEqual(t, res.StatusCode, http.StatusInternalServerError)

			lines := strings.Split(buf.String(), "\n")
			for _, line := range lines {
				if strings.HasPrefix(strings.TrimSpace(line), "->") {
					return
				}
			}
			t.Fatal("First func call line should start with ->.")
		})
	}

}

func assertEqual(t *testing.T, a, b any) {
	t.Helper()
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("expecting values to be equal but got: '%v' and '%v'", a, b)
	}
}
