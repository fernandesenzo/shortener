package shortener

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fernandesenzo/shortener/internal/domain"
)

type shortenLinkRequest struct {
	URL string `json:"url"`
}
type Handler struct {
	srv *Service
}

func NewHandler(srv *Service) *Handler {
	return &Handler{
		srv: srv,
	}
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenLinkRequest

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	link, err := h.srv.Shorten(r.Context(), req.URL)
	if err != nil {
		if errors.Is(err, domain.ErrLinkCreationFailed) {
			//TODO: log this error
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if errors.Is(err, domain.ErrURLTooLong) {
			http.Error(w, "url too long", http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, domain.ErrInvalidURL) {
			http.Error(w, "invalid url", http.StatusBadRequest)
			return
		}
		http.Error(w, "unknown error, try again", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(link); err != nil {
		//TODO: log this (i believe this only can happen if the connection drops
	}
}
