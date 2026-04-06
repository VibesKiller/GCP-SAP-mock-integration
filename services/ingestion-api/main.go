package main

import (
	"log"
	"net/http"

	"gcp-sap-mock-integration/internal/platform/logging"
	platformRuntime "gcp-sap-mock-integration/internal/platform/runtime"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := logging.New(cfg.ServiceName, cfg.Environment, cfg.LogLevel)
	application := newApp(cfg, logger)
	defer application.close()

	server := &http.Server{
		Addr:         cfg.address(),
		Handler:      application.routes(),
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
		IdleTimeout:  cfg.HTTPIdleTimeout,
	}

	ctx, cancel := platformRuntime.SignalContext()
	defer cancel()

	if err := platformRuntime.RunHTTPServer(ctx, logger, server); err != nil {
		logger.Error("ingestion-api stopped with error", "error", err)
		log.Fatalf("ingestion-api stopped with error: %v", err)
	}

	logger.Info("ingestion-api stopped gracefully")
}
