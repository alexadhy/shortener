package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/alexadhy/shortener/apiModel"
	"github.com/alexadhy/shortener/internal/log"
	"github.com/alexadhy/shortener/model"
	"github.com/alexadhy/shortener/persist"
	"github.com/alexadhy/shortener/render"
	"github.com/go-chi/chi/v5"
)

// API  is the name of the object that will handle all routes
type API struct {
	p                persist.Persist
	hostDomain       string
	domainFilterFunc func(string) bool
}

// New creates a new instance of the API
// hostDomain has to be in the form of {SCHEME}://{DOMAIN}.{TLD}
// domainFilterFn can be used to filter website we will shorten link to
func New(p persist.Persist, hostDomain string, domainFilterFn func(s string) bool) API {
	return API{p: p, hostDomain: hostDomain, domainFilterFunc: domainFilterFn}
}

// CreateShortLink will create short link from original URL
// will return the same shortened url if it already has one
func (a *API) CreateShortLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		handleErr(http.StatusBadRequest, errors.New("invalid request method"), w)
		return
	}

	defer r.Body.Close()

	var body apiModel.CreateShortLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		_, _ = render.Render(render.Response[any]{StatusCode: http.StatusBadRequest, Err: err}, w)
		return
	}

	u, err := url.Parse(body.OriginalURL)
	if err != nil {
		_, _ = render.Render(render.Response[any]{StatusCode: http.StatusBadRequest, Err: err}, w)
		return
	}

	if !a.domainFilterFunc(u.Host) {
		_, _ = render.Render(render.Response[any]{StatusCode: http.StatusBadRequest, Err: errors.New("non-whitelisted domain")}, w)
		return
	}

	shortData, err := model.New(body.OriginalURL)
	if err != nil {
		_, _ = render.Render(render.Response[any]{StatusCode: http.StatusBadRequest, Err: err}, w)
		return
	}

	if err := a.p.Set(r.Context(), shortData); err != nil {
		log.Errorf("CreateShortLink() Get: %v", err)
		handleErr(http.StatusInternalServerError, errors.New("internal error"), w)
		return
	}

	shortData, err = a.p.Get(r.Context(), shortData.Short)
	if err != nil {
		log.Errorf("CreateShortLink() Get: %v", err)
		handleErr(http.StatusInternalServerError, errors.New("internal error"), w)
		return
	}

	// construct short link url
	shortURL := a.hostDomain + "/" + shortData.Short
	_, _ = render.Render(
		render.Response[any]{
			StatusCode: http.StatusOK,
			Data:       apiModel.CreateShortLinkResponse{ShortLinkURL: shortURL},
		}, w,
	)
}

func (a *API) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		handleErr(http.StatusBadRequest, errors.New("invalid request method"), w)
		return
	}

	// get the shortened link
  key := chi.URLParam(r, "id")
	sd, err := a.p.Get(r.Context(), key)
	if err != nil {
		handleErr(http.StatusNotFound, errors.New("invalid link provider"), w)
		return
	}

	http.Redirect(w, r, sd.Orig, http.StatusMovedPermanently)
}

func handleErr(statusCode int, err error, w http.ResponseWriter) {
	_, _ = render.Render(
		render.Response[any]{StatusCode: statusCode, Err: err}, w,
	)
}
