package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type ResolveShortener interface {
	Shorten(raw string) (string, error)
	Resolve(id string) (string, error)
}

type Handler struct {
	BaseURL string
	rs      ResolveShortener
}

func New(baseURL string, s ResolveShortener) *Handler {
	return &Handler{BaseURL: baseURL, rs: s}
}

func (h *Handler) Shortify(c echo.Context) error {
	request := c.Request()
	if request.Method != echo.POST {
		return c.String(http.StatusBadRequest, "Invalid request method")
	}

	ct := request.Header.Get("Content-Type")
	if ct == "" || !strings.HasPrefix(strings.ToLower(ct), "text/plain") {
		return c.String(http.StatusBadRequest, "Content-Type must be text/plain")
	}

	request.Body = http.MaxBytesReader(c.Response().Writer, request.Body, 8<<10)
	defer request.Body.Close()
	raw, err := io.ReadAll(request.Body)
	if err != nil {
		return c.String(http.StatusBadRequest, "Failed to read body")
	}
	shortURL, err := h.rs.Shorten(string(raw))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	c.Response().Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Response().Writer.WriteHeader(http.StatusCreated)
	c.Response().Writer.Write([]byte(h.BaseURL + "/" + shortURL))

	return nil
}


func (h *Handler) Redirect(c echo.Context) error {
	request := c.Request()
	if request.Method != echo.GET {
		return c.String(http.StatusBadRequest, "Invalid request method")
	}

	id := c.Param("id")
	if id == "" {
		return c.String(http.StatusNotFound, "URL not found")
	}

	target, err := h.rs.Resolve(id)
	if err != nil {
		return c.String(http.StatusNotFound, "URL not found")
	}

	c.Redirect(http.StatusTemporaryRedirect, target)
	return nil
}
