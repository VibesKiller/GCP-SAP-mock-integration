package main

import (
	"fmt"
	"time"

	"gcp-sap-mock-integration/internal/platform/config"
	platformKafka "gcp-sap-mock-integration/internal/platform/kafka"
)

type appConfig struct {
	ServiceName        string
	Environment        string
	LogLevel           string
	Port               int
	HTTPReadTimeout    time.Duration
	HTTPWriteTimeout   time.Duration
	HTTPIdleTimeout    time.Duration
	KafkaBrokers       []string
	KafkaClientID      string
	KafkaConsumerGroup string
	KafkaTopics        []string
	KafkaDLQTopic      string
	RetryMaxAttempts   int
	RetryBackoff       time.Duration
	KafkaCommitTimeout time.Duration
	PostgresURL        string
}

func loadConfig() (appConfig, error) {
	port, err := config.GetInt("PORT", 8082)
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

	retryMaxAttempts, err := config.GetInt("PROCESSOR_RETRY_MAX_ATTEMPTS", 3)
	if err != nil {
		return appConfig{}, err
	}

	retryBackoff, err := config.GetDuration("PROCESSOR_RETRY_BACKOFF", 2*time.Second)
	if err != nil {
		return appConfig{}, err
	}

	commitTimeout, err := config.GetDuration("KAFKA_COMMIT_TIMEOUT", 5*time.Second)
	if err != nil {
		return appConfig{}, err
	}

	postgresURL, err := config.GetRequiredString("POSTGRES_URL")
	if err != nil {
		return appConfig{}, err
	}

	return appConfig{
		ServiceName:        config.GetString("SERVICE_NAME", "event-processor"),
		Environment:        config.GetString("ENVIRONMENT", "local"),
		LogLevel:           config.GetString("LOG_LEVEL", "info"),
		Port:               port,
		HTTPReadTimeout:    readTimeout,
		HTTPWriteTimeout:   writeTimeout,
		HTTPIdleTimeout:    idleTimeout,
		KafkaBrokers:       config.GetStringSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
		KafkaClientID:      config.GetString("KAFKA_CLIENT_ID", "event-processor"),
		KafkaConsumerGroup: config.GetString("KAFKA_CONSUMER_GROUP", platformKafka.ConsumerGroupEventProcessor),
		KafkaTopics: config.GetStringSlice("KAFKA_TOPICS", []string{
			platformKafka.TopicSalesOrders,
			platformKafka.TopicCustomers,
			platformKafka.TopicInvoices,
		}),
		KafkaDLQTopic:      config.GetString("KAFKA_DLQ_TOPIC", platformKafka.TopicIntegrationDLQ),
		RetryMaxAttempts:   retryMaxAttempts,
		RetryBackoff:       retryBackoff,
		KafkaCommitTimeout: commitTimeout,
		PostgresURL:        postgresURL,
	}, nil
}

func (c appConfig) address() string {
	return fmt.Sprintf(":%d", c.Port)
}
