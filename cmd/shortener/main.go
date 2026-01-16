package main

import (
	"log"

	"github.com/KKolyasik/url-shortify/internal/handler"
	"github.com/KKolyasik/url-shortify/internal/service"
	"github.com/KKolyasik/url-shortify/internal/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	st := storage.NewMemoryStore()
	svc := service.New(st)
	h := handler.New("http://localhost:8080", svc)

	router := gin.Default()
	router.POST("/", h.Shortify)
	router.GET("/:id", h.Redirect)
	err := router.Run()
	if err != nil {
		log.Fatal(err)
	}
}
