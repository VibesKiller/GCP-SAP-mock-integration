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
	PostgresURL      string
	DefaultPageSize  int
	MaxPageSize      int
}

func loadConfig() (appConfig, error) {
	port, err := config.GetInt("PORT", 8083)
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

	defaultPageSize, err := config.GetInt("DEFAULT_PAGE_SIZE", 20)
	if err != nil {
		return appConfig{}, err
	}

	maxPageSize, err := config.GetInt("MAX_PAGE_SIZE", 100)
	if err != nil {
		return appConfig{}, err
	}

	postgresURL, err := config.GetRequiredString("POSTGRES_URL")
	if err != nil {
		return appConfig{}, err
	}

	return appConfig{
		ServiceName:      config.GetString("SERVICE_NAME", "query-api"),
		Environment:      config.GetString("ENVIRONMENT", "local"),
		LogLevel:         config.GetString("LOG_LEVEL", "info"),
		Port:             port,
		HTTPReadTimeout:  readTimeout,
		HTTPWriteTimeout: writeTimeout,
		HTTPIdleTimeout:  idleTimeout,
		PostgresURL:      postgresURL,
		DefaultPageSize:  defaultPageSize,
		MaxPageSize:      maxPageSize,
	}, nil
}

func (c appConfig) address() string {
	return fmt.Sprintf(":%d", c.Port)
}
