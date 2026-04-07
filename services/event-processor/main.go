package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"gcp-sap-mock-integration/internal/platform/logging"
	platformRuntime "gcp-sap-mock-integration/internal/platform/runtime"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := logging.New(cfg.ServiceName, cfg.Environment, cfg.LogLevel)
	rootCtx, stop := platformRuntime.SignalContext()
	defer stop()

	application, err := newApp(rootCtx, cfg, logger)
	if err != nil {
		logger.Error("initialize event-processor", "error", err)
		os.Exit(1)
	}
	defer application.close()

	server := &http.Server{
		Addr:         cfg.address(),
		Handler:      application.routes(),
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
		IdleTimeout:  cfg.HTTPIdleTimeout,
	}

	runCtx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	errCh := make(chan error, 2)
	go func() {
		errCh <- platformRuntime.RunHTTPServer(runCtx, logger, server)
	}()
	go func() {
		errCh <- application.run(runCtx)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			cancel()
			logger.Error("event-processor stopped with error", "error", err)
			os.Exit(1)
		}
	case <-runCtx.Done():
	}

	logger.Info("event-processor stopped gracefully")
}
