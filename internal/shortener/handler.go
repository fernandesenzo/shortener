package shortener

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/fernandesenzo/shortener/internal/domain"
)

type Handler struct {
	srv *Service
}

func NewHandler(srv *Service) *Handler {
	return &Handler{
		srv: srv,
	}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "invalid code", http.StatusBadRequest)
		return
	}
	link, err := h.srv.Get(r.Context(), code)
	if err != nil {
		if errors.Is(err, domain.ErrLinkNotFound) {
			http.Error(w, "link not found or expired", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		slog.ErrorContext(r.Context(), "failed to get link", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := getLinkResponse{
		URL: link.OriginalURL,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.ErrorContext(r.Context(), "failed to encode response", "error", err)
	}
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		http.Error(w, "unsupported content type", http.StatusUnsupportedMediaType)
		return
	}

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
			slog.ErrorContext(r.Context(), "failed to create link", "error", err, "url", req.URL)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		slog.WarnContext(r.Context(), "user sent bad request", "error", err, "url", req.URL)
		if errors.Is(err, domain.ErrURLTooLong) {
			http.Error(w, "url too long", http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, domain.ErrInvalidURL) {
			http.Error(w, "invalid url", http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := shortenLinkResponse{
		Code: link.Code,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.ErrorContext(r.Context(), "failed to encode response", "error", err)
	}
}
