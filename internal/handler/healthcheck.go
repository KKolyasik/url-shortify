package handler

import (
	"context"
	"net/http"
)


type Pinger interface {
	Ping(ctx context.Context) error
}

type HealthCheckHandler struct {
	DB Pinger
}

func NewHealthCheckHandler(db Pinger) *HealthCheckHandler {
	return &HealthCheckHandler{
		DB: db,
	}
}

func (h *HealthCheckHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.DB.Ping(r.Context()); err != nil {
		http.Error(w, "DB connect error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}