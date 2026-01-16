package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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

func (h *Handler) Shortify(c *gin.Context) {
	request := c.Request

	ct := request.Header.Get("Content-Type")
	if ct == "" || !strings.HasPrefix(strings.ToLower(ct), "text/plain") {
		c.String(http.StatusBadRequest, "Content-Type must be text/plain")
		return
	}

	request.Body = http.MaxBytesReader(c.Writer, request.Body, 8<<10)
	defer request.Body.Close()
	raw, err := io.ReadAll(request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "Failed to read body")
		return
	}
	shortURL, err := h.rs.Shorten(string(raw))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(http.StatusCreated)
	c.Writer.Write([]byte(h.BaseURL + "/" + shortURL))

}


func (h *Handler) Redirect(c *gin.Context) {

	id := c.Param("id")
	if id == "" {
		c.String(http.StatusNotFound, "URL not found")
		return
	}

	target, err := h.rs.Resolve(id)
	if err != nil {
		c.String(http.StatusNotFound, "URL not found")
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, target)
}
