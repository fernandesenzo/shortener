package user

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
	return &Handler{srv: srv}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		h.sendError(w, r, "unsupported content type", http.StatusUnsupportedMediaType)
		return
	}

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, r, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.srv.Create(r.Context(), req.Nickname, req.Password)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	h.sendJSON(w, r, http.StatusCreated, createResponse{
		ID:        user.ID,
		Nickname:  user.Nickname,
		CreatedAt: user.CreatedAt.String(),
	})
}

func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrNicknameAlreadyUsed):
		h.sendError(w, r, "nickname already in use", http.StatusConflict)
	case errors.Is(err, domain.ErrPasswordTooShort), errors.Is(err, domain.ErrPasswordTooLong):
		h.sendError(w, r, "invalid password length", http.StatusUnprocessableEntity)
	default:
		h.sendError(w, r, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) sendJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.ErrorContext(r.Context(), "failed to encode json response", "error", err)
	}
}

func (h *Handler) sendError(w http.ResponseWriter, r *http.Request, msg string, status int) {
	h.sendJSON(w, r, status, map[string]string{"error": msg})
}
