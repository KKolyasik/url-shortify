package handler

import (
	"io"
	"net/http"
	"strings"
)

type Shortener interface {
	Shorten(raw string) (string, error)
	Resolve(id string) (string, error)
}

type Handler struct {
	BaseURL string
	S       Shortener
}

func New(baseURL string, s Shortener) *Handler {
	return &Handler{BaseURL: baseURL, S: s}
}

func (h *Handler) Shortify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	ct := r.Header.Get("Content-Type")
	if ct == "" || !strings.HasPrefix(strings.ToLower(ct), "text/plain") {
		http.Error(w, "Content-Type must be text/plain", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	defer r.Body.Close()
	raw, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	shortURL, err := h.S.Shorten(string(raw))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(h.BaseURL + "/" + shortURL))
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	target, err := h.S.Resolve(id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}
