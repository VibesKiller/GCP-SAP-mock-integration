package main

import (
	"fmt"
	"time"

	"gcp-sap-mock-integration/internal/platform/config"
)

type appConfig struct {
	ServiceName      string
	Environment      string
	LogLevel         string
	Port             int
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	HTTPIdleTimeout  time.Duration
	AutoDispatch     bool
	IngestionBaseURL string
	DispatchTimeout  time.Duration
}

func loadConfig() (appConfig, error) {
	port, err := config.GetInt("PORT", 8080)
	if err != nil {
		return appConfig{}, err
	}

	readTimeout, err := config.GetDuration("HTTP_READ_TIMEOUT", 5*time.Second)
	if err != nil {
		return appConfig{}, err
	}

	writeTimeout, err := config.GetDuration("HTTP_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return appConfig{}, err
	}

	idleTimeout, err := config.GetDuration("HTTP_IDLE_TIMEOUT", 30*time.Second)
	if err != nil {
		return appConfig{}, err
	}

	autoDispatch, err := config.GetBool("AUTO_DISPATCH", false)
	if err != nil {
		return appConfig{}, err
	}

	dispatchTimeout, err := config.GetDuration("DISPATCH_TIMEOUT", 5*time.Second)
	if err != nil {
		return appConfig{}, err
	}

	return appConfig{
		ServiceName:      config.GetString("SERVICE_NAME", "sap-mock-api"),
		Environment:      config.GetString("ENVIRONMENT", "local"),
		LogLevel:         config.GetString("LOG_LEVEL", "info"),
		Port:             port,
		HTTPReadTimeout:  readTimeout,
		HTTPWriteTimeout: writeTimeout,
		HTTPIdleTimeout:  idleTimeout,
		AutoDispatch:     autoDispatch,
		IngestionBaseURL: config.GetString("INGESTION_API_BASE_URL", "http://localhost:8081"),
		DispatchTimeout:  dispatchTimeout,
	}, nil
}

func (c appConfig) address() string {
	return fmt.Sprintf(":%d", c.Port)
}
