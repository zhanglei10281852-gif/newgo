package main

import (
	"context"
	"github.com/zhanglei10281852-gif/newgo/internal/clock"
	"github.com/zhanglei10281852-gif/newgo/internal/config"
	"github.com/zhanglei10281852-gif/newgo/internal/httpapi"
	"github.com/zhanglei10281852-gif/newgo/internal/service"
	"github.com/zhanglei10281852-gif/newgo/internal/storage/sqlite"
	"github.com/zhanglei10281852-gif/newgo/internal/worker"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()
	store, err := sqlite.Open(ctx, cfg.DatabasePath)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	services := service.New(store, clock.System{}, cfg.SessionTTL, cfg.PermitTTL)
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: (httpapi.Server{Services: services, Logger: logger}).Handler(), ReadHeaderTimeout: 5 * time.Second}
	workerCtx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	go func() {
		_ = (worker.Worker{Store: store, Clock: clock.System{}, Interval: cfg.WorkerInterval, BatchSize: cfg.WorkerBatchSize, Logger: logger}).Run(workerCtx)
	}()
	go func() {
		logger.Info("server started", "addr", cfg.HTTPAddr)
		if e := server.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			logger.Error("server stopped", "error", e)
		}
	}()
	<-workerCtx.Done()
	shutdown, stop := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer stop()
	_ = server.Shutdown(shutdown)
}
