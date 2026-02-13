package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/KKolyasik/url-shortify/internal/model"
)

type ResolveShortener interface {
	Shorten(ctx context.Context, raw string) (string, error)
	Resolve(ctx context.Context, id string) (string, error)
}

type Handler struct {
	BaseURL string
	rs      ResolveShortener
}

func New(baseURL string, s ResolveShortener) *Handler {
	return &Handler{BaseURL: baseURL, rs: s}
}

func (h *Handler) Shortify(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
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
	shortURL, err := h.rs.Shorten(r.Context(), string(raw))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(h.BaseURL + "/" + shortURL))

}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request, id string) {

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if id == "" {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	target, err := h.rs.Resolve(r.Context(), id)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
}

func (h *Handler) ShortifyJSON(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ct := r.Header.Get("Content-Type")
	if ct == "" || !strings.EqualFold(ct, "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	defer r.Body.Close()

	var u model.URLRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&u); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Cannot decode request JSON body", http.StatusBadRequest)
		return
	}

	shortURL, err := h.rs.Shorten(r.Context(), u.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := model.URLResponse{Result: h.BaseURL + "/" + shortURL}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(resp); err != nil {
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}

}

func (h *Handler) ShortenBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	ct := r.Header.Get("Content-Type")
	if ct == "" || !strings.EqualFold(ct, "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	defer r.Body.Close()

	var batches []model.URLBatchRequest
	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&batches); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Cannot decode request JSON body", http.StatusBadRequest)
		return
	}

	var response []model.URLBatchResponse
	for _, batch := range batches {
		var resp model.URLBatchResponse
		shortURL, err := h.rs.Shorten(r.Context(), batch.OriginalURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp.CorrelationId = batch.CorrelationId
		resp.ShortURL = h.BaseURL + "/" + shortURL

		response = append(response, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	enc := json.NewEncoder(w)
	if err := enc.Encode(response); err != nil {
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}
