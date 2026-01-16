package main

import (
	"log"

	"github.com/KKolyasik/url-shortify/internal/handler"
	"github.com/KKolyasik/url-shortify/internal/service"
	"github.com/KKolyasik/url-shortify/internal/storage"
	"github.com/labstack/echo/v4"
)

func main() {
	st := storage.NewMemoryStore()
	svc := service.New(st)
	h := handler.New("http://localhost:8080", svc)

	e := echo.New()
	e.POST("/", h.Shortify)
	e.GET("/:id", h.Redirect)
	err := e.Start(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
