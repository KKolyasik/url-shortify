package main

import (
	"compress/gzip"
	"log"

	"github.com/KKolyasik/url-shortify/internal/config"
	"github.com/KKolyasik/url-shortify/internal/encoding"
	"github.com/KKolyasik/url-shortify/internal/handler"
	"github.com/KKolyasik/url-shortify/internal/logger"
	"github.com/KKolyasik/url-shortify/internal/service"
	"github.com/KKolyasik/url-shortify/internal/storage"
	"github.com/KKolyasik/url-shortify/internal/transport"
	"github.com/KKolyasik/url-shortify/internal/transport/ginmw"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	logger.Initialize("Info")
	defer logger.Log.Sync()
	en := encoding.NewGzip(gzip.BestCompression)
	st := storage.NewMemoryStore()
	svc := service.New(st)
	h := handler.New(cfg.URLAddr, svc)
	router := gin.Default()
	router.Use(ginmw.RequestLogger(logger.Log.Sugar()))
	router.Use(gin.Recovery())
	router.Use(ginmw.GinContentEncoding(logger.Log.Sugar(), en))
	router.POST("/api/shorten", transport.GinShortifyJSON(h))
	router.POST("/", transport.GinShortify(h))
	router.GET("/:id", transport.GinRedirect(h))
	err = router.Run(cfg.Addr.String())
	if err != nil {
		log.Fatal(err)
	}
}
