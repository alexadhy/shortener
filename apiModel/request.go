package apiModel

// CreateShortLinkRequest is the request type to create new short link URL
type CreateShortLinkRequest struct {
	OriginalURL string `json:"url"`
}

// CreateShortLinkResponse is the response type to create new short link URL
type CreateShortLinkResponse struct {
	ShortLinkURL string `json:"url"`
}
