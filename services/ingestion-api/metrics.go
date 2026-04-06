package main

import "github.com/prometheus/client_golang/prometheus/promauto"
import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	requestsTotal   *prometheus.CounterVec
	publishedTotal  *prometheus.CounterVec
	publishDuration *prometheus.HistogramVec
}

func newMetrics() *metrics {
	return &metrics{
		requestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sap_integration",
			Subsystem: "ingestion_api",
			Name:      "requests_total",
			Help:      "Total number of ingestion requests by event type and result.",
		}, []string{"event_type", "result"}),
		publishedTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sap_integration",
			Subsystem: "ingestion_api",
			Name:      "published_total",
			Help:      "Total number of published Kafka messages by topic and result.",
		}, []string{"topic", "event_type", "result"}),
		publishDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "sap_integration",
			Subsystem: "ingestion_api",
			Name:      "publish_duration_seconds",
			Help:      "Kafka publish duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"topic", "event_type"}),
	}
}
