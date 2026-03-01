package shortener

type shortenLinkRequest struct {
	URL string `json:"url"`
}

type shortenLinkResponse struct {
	Code string `json:"code"`
}

type deleteLinkResponse struct {
	Message string `json:"message"`
}
