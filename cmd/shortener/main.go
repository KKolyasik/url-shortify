package main

import (
	"compress/gzip"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KKolyasik/url-shortify/internal/audit"
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

type noopPinger struct{}

func (p *noopPinger) Ping(_ context.Context) error { return nil }

func main() {
	logger.Initialize("Info")
	defer logger.Log.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Log.Sugar().Fatal(err.Error())
	}

	en := encoding.NewGzip(gzip.BestCompression)

	var (
		st     service.Storage
		fs     *storage.FileStorage
		pinger handler.Pinger
	)

	switch {
	case cfg.DBURL != "":
		logger.Log.Info("Подключаемся к БД")
		db := storage.NewPostgresDB(context.Background(), cfg.DBURL)
		st = db
		pinger = db
	case cfg.FileStorage != "":
		logger.Log.Info("Используем файл для хранения ссылок")
		memoryStore := storage.NewMemoryStore()
		st = memoryStore
		fs, err = storage.NewFileStorage(cfg.FileStorage, memoryStore)
		pinger = &noopPinger{}
		if err != nil {
			logger.Log.Sugar().Fatal(err.Error())
		}
		err = fs.Restore(context.Background())
		if err != nil {
			logger.Log.Sugar().Error("Не удалось восстановить данные из файла")
		}
		defer fs.Close()
	default:
		logger.Log.Info("Используем только in-memory хранилище")
		memoryStore := storage.NewMemoryStore()
		st = memoryStore
		pinger = &noopPinger{}
	}

	if fs != nil {
		defer func() {
			saveCtx, saveCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer saveCancel()
			err := fs.Save(saveCtx)
			if err != nil {
				logger.Log.Sugar().Fatal(err.Error())
			}
		}()
	}

	appCtx, appCancel := context.WithCancel(context.Background())
	svc := service.New(appCtx, st)
	auditor := audit.NewAudit(&cfg.Audit, logger.Log)

	h := handler.New(cfg.URLAddr, svc, auditor)
	hch := handler.NewHealthCheckHandler(pinger)

	router := gin.Default()
	router.Use(ginmw.RequestLogger(logger.Log.Sugar()))
	router.Use(gin.Recovery())
	router.Use(ginmw.GinContentEncoding(logger.Log.Sugar(), en))
	router.Use(ginmw.GinAuthorization(&cfg))
	router.POST("/api/shorten", transport.GinShortifyJSON(h))
	router.POST("/api/shorten/batch", transport.GinShortifyBatch(h))
	router.POST("/", transport.GinShortify(h))
	router.GET("/:id", transport.GinRedirect(h))
	router.GET("/api/user/urls", transport.GinUserURLS(h))
	router.DELETE("/api/user/urls", transport.GinDeleteURL(h))

	router.GET("/ping", gin.WrapF(hch.HealthCheck))

	srv := &http.Server{
		Addr:              cfg.Addr.String(),
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*5)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Sugar().Fatal("Принудительное завершение сервера")
	}
	appCancel()
	svc.Stop()

	logger.Log.Sugar().Info("Сервер завершил работу")
}
