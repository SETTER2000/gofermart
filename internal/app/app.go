package app

import (
	"fmt"
	"github.com/SETTER2000/gofermart/internal/client"
	"github.com/SETTER2000/gofermart/scripts"
	"github.com/go-chi/chi/v5/middleware"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SETTER2000/gofermart/config"
	v1 "github.com/SETTER2000/gofermart/internal/controller/http/v1"
	"github.com/SETTER2000/gofermart/internal/server"
	"github.com/SETTER2000/gofermart/internal/usecase"
	"github.com/SETTER2000/gofermart/internal/usecase/repo"
	"github.com/SETTER2000/gofermart/pkg/log/logger"
	"github.com/go-chi/chi/v5"
	"github.com/xlab/closer"
)

func Run(cfg *config.Config) {
	closer.Bind(cleanup)
	// logger
	l := logger.New(cfg.Log.Level)
	// seed
	rand.Seed(time.Now().UnixNano())
	// Use case
	var gofermartUseCase usecase.Gofermart
	if !scripts.CheckEnvironFlag("DATABASE_URI", cfg.Storage.ConnectDB) {
		if cfg.FileStorage == "" {
			l.Warn("In memory storage!!!")
			gofermartUseCase = usecase.New(repo.NewInMemory(cfg))
		} else {
			l.Info("File storage - is work...")
			gofermartUseCase = usecase.New(repo.NewInFiles(cfg))
			if err := gofermartUseCase.ReadService(); err != nil {
				l.Error(fmt.Errorf("app - Read - gofermartUseCase.ReadService: %w", err))
			}
		}
	} else {
		l.Info("DB SQL - is work...")
		gofermartUseCase = usecase.New(repo.NewInSQL(cfg))
	}

	// HTTP AClient - клиент для accrual сервиса
	client := client.NewAClient(gofermartUseCase, cfg)
	client.Start()
	// HTTP Server
	handler := chi.NewRouter()
	// GZIP
	handler.Use(middleware.AllowContentEncoding("deflate", "gzip"))
	// Router
	v1.NewRouter(handler, l, gofermartUseCase, cfg, client)
	httpServer := server.New(handler, server.Host(cfg.HTTP.ServerAddress))

	// waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		if err := gofermartUseCase.SaveService(); err != nil {
			l.Error(fmt.Errorf("app - Save - gofermartUseCase.SaveService: %w", err))
		}
		l.Info("app - Run - signal: " + s.String())
	case err := <-httpServer.Notify():
		l.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	closer.Hold()

	err := httpServer.Shutdown()
	if err != nil {
		l.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}

func cleanup() {
	fmt.Println("Hang on! I'm closing some DBs, wiping some trails..")
	time.Sleep(1 * time.Second)
	fmt.Println("  Done...")
}
