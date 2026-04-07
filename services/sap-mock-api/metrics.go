package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	simulationsTotal *prometheus.CounterVec
	dispatchTotal    *prometheus.CounterVec
	dispatchDuration *prometheus.HistogramVec
}

func newMetrics() *metrics {
	return &metrics{
		simulationsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sap_integration",
			Subsystem: "sap_mock_api",
			Name:      "simulations_total",
			Help:      "Total number of mock SAP simulation requests by event type and result.",
		}, []string{"event_type", "result"}),
		dispatchTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sap_integration",
			Subsystem: "sap_mock_api",
			Name:      "dispatch_total",
			Help:      "Total number of mock SAP dispatch attempts to ingestion-api by event type and result.",
		}, []string{"event_type", "result"}),
		dispatchDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "sap_integration",
			Subsystem: "sap_mock_api",
			Name:      "dispatch_duration_seconds",
			Help:      "Time spent dispatching mock SAP payloads to ingestion-api.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"event_type"}),
	}
}
