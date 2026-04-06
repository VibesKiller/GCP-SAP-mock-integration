package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	consumedTotal     *prometheus.CounterVec
	duplicatesTotal   prometheus.Counter
	retriesTotal      *prometheus.CounterVec
	dlqTotal          *prometheus.CounterVec
	processingLatency *prometheus.HistogramVec
}

func newMetrics() *metrics {
	return &metrics{
		consumedTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sap_integration",
			Subsystem: "event_processor",
			Name:      "consumed_total",
			Help:      "Total number of consumed Kafka messages by topic and result.",
		}, []string{"topic", "result"}),
		duplicatesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: "sap_integration",
			Subsystem: "event_processor",
			Name:      "duplicate_events_total",
			Help:      "Total number of duplicate events ignored by idempotency checks.",
		}),
		retriesTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sap_integration",
			Subsystem: "event_processor",
			Name:      "retries_total",
			Help:      "Total number of retry attempts by event type.",
		}, []string{"event_type"}),
		dlqTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sap_integration",
			Subsystem: "event_processor",
			Name:      "dlq_total",
			Help:      "Total number of messages written to the dead-letter topic by reason.",
		}, []string{"reason"}),
		processingLatency: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "sap_integration",
			Subsystem: "event_processor",
			Name:      "processing_duration_seconds",
			Help:      "Time spent processing events by event type.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"event_type"}),
	}
}
