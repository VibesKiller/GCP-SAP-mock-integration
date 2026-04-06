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
	application, err := newApp(cfg, logger)
	if err != nil {
		logger.Error("initialize query-api", "error", err)
		log.Fatalf("initialize query-api: %v", err)
	}
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
		logger.Error("query-api stopped with error", "error", err)
		log.Fatalf("query-api stopped with error: %v", err)
	}

	logger.Info("query-api stopped gracefully")
}
