package main

import (
	"fmt"
	"time"

	"gcp-sap-mock-integration/internal/platform/config"
	platformKafka "gcp-sap-mock-integration/internal/platform/kafka"
)

type appConfig struct {
	ServiceName       string
	Environment       string
	LogLevel          string
	Port              int
	HTTPReadTimeout   time.Duration
	HTTPWriteTimeout  time.Duration
	HTTPIdleTimeout   time.Duration
	Kafka             platformKafka.ClientConfig
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

	kafkaTLSEnabled, err := config.GetBool("KAFKA_TLS_ENABLED", false)
	if err != nil {
		return appConfig{}, err
	}

	kafkaTLSInsecureSkipVerify, err := config.GetBool("KAFKA_TLS_INSECURE_SKIP_VERIFY", false)
	if err != nil {
		return appConfig{}, err
	}

	kafkaConfig := platformKafka.ClientConfig{
		Brokers:               config.GetStringSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		ClientID:              config.GetString("KAFKA_CLIENT_ID", "ingestion-api"),
		TLSEnabled:            kafkaTLSEnabled,
		TLSInsecureSkipVerify: kafkaTLSInsecureSkipVerify,
		TLSServerName:         config.GetString("KAFKA_TLS_SERVER_NAME", ""),
		TLSCAFile:             config.GetString("KAFKA_TLS_CA_CERT_FILE", ""),
		AuthMode:              platformKafka.AuthMode(config.GetString("KAFKA_AUTH_MODE", string(platformKafka.AuthModeNone))),
		SASLUsername:          config.GetString("KAFKA_SASL_USERNAME", ""),
		SASLPassword:          config.GetString("KAFKA_SASL_PASSWORD", ""),
		GCPPrincipalEmail:     config.GetString("KAFKA_GCP_PRINCIPAL_EMAIL", ""),
		GCPAccessTokenScope:   config.GetString("KAFKA_GCP_ACCESS_TOKEN_SCOPE", platformKafka.DefaultGoogleAccessTokenScope),
	}
	if err := kafkaConfig.Validate(); err != nil {
		return appConfig{}, fmt.Errorf("validate Kafka client config: %w", err)
	}

	return appConfig{
		ServiceName:       config.GetString("SERVICE_NAME", "ingestion-api"),
		Environment:       config.GetString("ENVIRONMENT", "local"),
		LogLevel:          config.GetString("LOG_LEVEL", "info"),
		Port:              port,
		HTTPReadTimeout:   readTimeout,
		HTTPWriteTimeout:  writeTimeout,
		HTTPIdleTimeout:   idleTimeout,
		Kafka:             kafkaConfig,
		KafkaWriteTimeout: kafkaWriteTimeout,
		SourceSystem:      config.GetString("INGESTION_SOURCE_SYSTEM", "sap-s4hana"),
	}, nil
}

func (c appConfig) address() string {
	return fmt.Sprintf(":%d", c.Port)
}
