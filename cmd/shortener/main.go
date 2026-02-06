package main

import (
	"compress/gzip"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	logger.Initialize("Info")
	defer logger.Log.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Log.Sugar().Fatal(err.Error())
	}

	en := encoding.NewGzip(gzip.BestCompression)

	st := storage.NewMemoryStore()
	db := storage.NewPostgresDB(cfg.DBURL)

	fs, err := storage.NewFileStorage(cfg.FileStorage, st)
	if err != nil {
		logger.Log.Sugar().Fatal(err.Error())
	}
	defer func ()  {
		err := fs.Close()
		if err != nil {
			logger.Log.Sugar().Fatal(err.Error())
		}
	}()

	if err := fs.Restore(context.Background()); err != nil {
		logger.Log.Sugar().Fatal(err.Error())
	}
	defer func ()  {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := fs.Save(ctx)
		if err != nil {
			logger.Log.Sugar().Fatal(err.Error())
		}
	}()

	svc := service.New(st)

	h := handler.New(cfg.URLAddr, svc)
	hch := handler.NewHealthCheckHandler(db)

	router := gin.Default()
	router.Use(ginmw.RequestLogger(logger.Log.Sugar()))
	router.Use(gin.Recovery())
	router.Use(ginmw.GinContentEncoding(logger.Log.Sugar(), en))
	router.POST("/api/shorten", transport.GinShortifyJSON(h))
	router.POST("/", transport.GinShortify(h))
	router.GET("/:id", transport.GinRedirect(h))

	router.GET("/ping", gin.WrapF(hch.HealthCheck))

	srv := &http.Server{
		Addr:    cfg.Addr.String(),
		Handler: router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout: cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout: cfg.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Sugar().Fatal(err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Sugar().Info("Завершение работы сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Sugar().Fatal("Принудительное завершение сервера")
	}

	logger.Log.Sugar().Info("Сервер завершил работу")
}
