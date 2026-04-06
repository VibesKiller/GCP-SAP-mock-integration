package main

import (
	"fmt"
	"time"

	"gcp-sap-mock-integration/internal/platform/config"
)

type appConfig struct {
	ServiceName       string
	Environment       string
	LogLevel          string
	Port              int
	HTTPReadTimeout   time.Duration
	HTTPWriteTimeout  time.Duration
	HTTPIdleTimeout   time.Duration
	KafkaBrokers      []string
	KafkaClientID     string
	KafkaWriteTimeout time.Duration
	SourceSystem      string
}

func loadConfig() (appConfig, error) {
	port, err := config.GetInt("PORT", 8081)
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

	kafkaWriteTimeout, err := config.GetDuration("KAFKA_WRITE_TIMEOUT", 5*time.Second)
	if err != nil {
		return appConfig{}, err
	}

	return appConfig{
		ServiceName:       config.GetString("SERVICE_NAME", "ingestion-api"),
		Environment:       config.GetString("ENVIRONMENT", "local"),
		LogLevel:          config.GetString("LOG_LEVEL", "info"),
		Port:              port,
		HTTPReadTimeout:   readTimeout,
		HTTPWriteTimeout:  writeTimeout,
		HTTPIdleTimeout:   idleTimeout,
		KafkaBrokers:      config.GetStringSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		KafkaClientID:     config.GetString("KAFKA_CLIENT_ID", "ingestion-api"),
		KafkaWriteTimeout: kafkaWriteTimeout,
		SourceSystem:      config.GetString("INGESTION_SOURCE_SYSTEM", "sap-s4hana"),
	}, nil
}

func (c appConfig) address() string {
	return fmt.Sprintf(":%d", c.Port)
}
