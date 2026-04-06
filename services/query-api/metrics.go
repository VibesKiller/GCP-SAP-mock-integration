package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type metrics struct {
	requestsTotal *prometheus.CounterVec
}

func newMetrics() *metrics {
	return &metrics{
		requestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "sap_integration",
			Subsystem: "query_api",
			Name:      "requests_total",
			Help:      "Total number of query API requests by endpoint and result.",
		}, []string{"endpoint", "result"}),
	}
}
