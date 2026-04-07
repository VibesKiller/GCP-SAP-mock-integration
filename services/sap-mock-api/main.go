package main

import (
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

	samples, err := loadSamples()
	if err != nil {
		logger.Error("load sample catalog", "error", err)
		os.Exit(1)
	}

	application := newApp(cfg, logger, samples)
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
		logger.Error("sap-mock-api stopped with error", "error", err)
		os.Exit(1)
	}

	logger.Info("sap-mock-api stopped gracefully")
}
