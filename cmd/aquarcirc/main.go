package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aquarecirc/internal/service"
)

func main() {
	logger := log.New(os.Stdout, "aquarcirc ", log.LstdFlags)
	cfg := service.LoadConfig()
	svc, err := service.New(cfg, logger)
	if err != nil {
		logger.Fatalf("init service: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go svc.RunLoops(ctx)
	server := svc.HTTPServer()
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("serve: %v", err)
		}
	}()
	if err := svc.WaitReady(10 * time.Second); err != nil {
		logger.Fatalf("service not ready: %v", err)
	}
	if err := svc.HTTPProbe("127.0.0.1" + cfg.Addr); err != nil {
		logger.Fatalf("http probe: %v", err)
	}
	logger.Printf("aquarcirc ready on %s", cfg.Addr)
	go waitForShutdown(server, cancel)
	<-ctx.Done()
}

func waitForShutdown(server *http.Server, cancel context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)
}
