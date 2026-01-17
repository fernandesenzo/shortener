package shortener

import (
	"encoding/json"
	"errors"
	"net/http"
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

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	link, err := h.srv.Shorten(req.URL)
	if err != nil {
		if errors.Is(err, ErrLinkNotSaved) || errors.Is(err, ErrGenCode) {
			//TODO: log this error
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(link); err != nil {
		//TODO: log this (i believe this only can happen if the connection drops
	}
}
