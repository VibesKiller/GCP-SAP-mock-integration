package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	platformHttp "gcp-sap-mock-integration/internal/platform/httpx"
	platformPostgres "gcp-sap-mock-integration/internal/platform/postgres"
)

type app struct {
	config     appConfig
	logger     *slog.Logger
	repository *repository
	metrics    *metrics
}

func newApp(cfg appConfig, logger *slog.Logger) (*app, error) {
	db, err := platformPostgres.NewPool(context.Background(), cfg.PostgresURL)
	if err != nil {
		return nil, err
	}

	return &app{
		config:     cfg,
		logger:     logger,
		repository: newRepository(db),
		metrics:    newMetrics(),
	}, nil
}

func (a *app) close() {
	a.repository.db.Close()
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	platformHttp.RegisterHealthEndpoints(mux, a.config.ServiceName, a.repository.ready)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("GET /api/v1/customers", a.handleListCustomers)
	mux.HandleFunc("GET /api/v1/customers/{customerID}", a.handleGetCustomer)
	mux.HandleFunc("GET /api/v1/orders", a.handleListOrders)
	mux.HandleFunc("GET /api/v1/orders/{orderID}", a.handleGetOrder)
	mux.HandleFunc("GET /api/v1/invoices", a.handleListInvoices)
	mux.HandleFunc("GET /api/v1/invoices/{invoiceID}", a.handleGetInvoice)

	return platformHttp.Chain(mux,
		platformHttp.CorrelationMiddleware(),
		platformHttp.RecoveryMiddleware(a.logger),
		platformHttp.LoggingMiddleware(a.logger),
	)
}

func (a *app) handleListCustomers(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := a.parsePagination(r)
	if err != nil {
		a.metrics.requestsTotal.WithLabelValues("customers_list", "invalid").Inc()
		platformHttp.WriteError(w, http.StatusBadRequest, err.Error(), platformHttp.CorrelationIDFromContext(r.Context()))
		return
	}

	result, err := a.repository.listCustomers(r.Context(), limit, offset, customerFilters{
		Status:      r.URL.Query().Get("status"),
		CountryCode: r.URL.Query().Get("country_code"),
		Segment:     r.URL.Query().Get("segment"),
	})
	if err != nil {
		a.metrics.requestsTotal.WithLabelValues("customers_list", "error").Inc()
		platformHttp.WriteError(w, http.StatusInternalServerError, err.Error(), platformHttp.CorrelationIDFromContext(r.Context()))
		return
	}

	a.metrics.requestsTotal.WithLabelValues("customers_list", "success").Inc()
	platformHttp.WriteJSON(w, http.StatusOK, result)
}

func (a *app) handleGetCustomer(w http.ResponseWriter, r *http.Request) {
	result, err := a.repository.getCustomer(r.Context(), r.PathValue("customerID"))
	if err != nil {
		if notFoundError(err) {
			a.metrics.requestsTotal.WithLabelValues("customers_get", "not_found").Inc()
			platformHttp.WriteError(w, http.StatusNotFound, "customer not found", platformHttp.CorrelationIDFromContext(r.Context()))
			return
		}
		a.metrics.requestsTotal.WithLabelValues("customers_get", "error").Inc()
		platformHttp.WriteError(w, http.StatusInternalServerError, err.Error(), platformHttp.CorrelationIDFromContext(r.Context()))
		return
	}

	a.metrics.requestsTotal.WithLabelValues("customers_get", "success").Inc()
	platformHttp.WriteJSON(w, http.StatusOK, result)
}

func (a *app) handleListOrders(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := a.parsePagination(r)
	if err != nil {
		a.metrics.requestsTotal.WithLabelValues("orders_list", "invalid").Inc()
		platformHttp.WriteError(w, http.StatusBadRequest, err.Error(), platformHttp.CorrelationIDFromContext(r.Context()))
		return
	}

	result, err := a.repository.listOrders(r.Context(), limit, offset, orderFilters{
		CustomerID: r.URL.Query().Get("customer_id"),
		Status:     r.URL.Query().Get("status"),
	})
	if err != nil {
		a.metrics.requestsTotal.WithLabelValues("orders_list", "error").Inc()
		platformHttp.WriteError(w, http.StatusInternalServerError, err.Error(), platformHttp.CorrelationIDFromContext(r.Context()))
		return
	}

	a.metrics.requestsTotal.WithLabelValues("orders_list", "success").Inc()
	platformHttp.WriteJSON(w, http.StatusOK, result)
}

func (a *app) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	result, err := a.repository.getOrder(r.Context(), r.PathValue("orderID"))
	if err != nil {
		if notFoundError(err) {
			a.metrics.requestsTotal.WithLabelValues("orders_get", "not_found").Inc()
			platformHttp.WriteError(w, http.StatusNotFound, "order not found", platformHttp.CorrelationIDFromContext(r.Context()))
			return
		}
		a.metrics.requestsTotal.WithLabelValues("orders_get", "error").Inc()
		platformHttp.WriteError(w, http.StatusInternalServerError, err.Error(), platformHttp.CorrelationIDFromContext(r.Context()))
		return
	}

	a.metrics.requestsTotal.WithLabelValues("orders_get", "success").Inc()
	platformHttp.WriteJSON(w, http.StatusOK, result)
}

func (a *app) handleListInvoices(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := a.parsePagination(r)
	if err != nil {
		a.metrics.requestsTotal.WithLabelValues("invoices_list", "invalid").Inc()
		platformHttp.WriteError(w, http.StatusBadRequest, err.Error(), platformHttp.CorrelationIDFromContext(r.Context()))
		return
	}

	result, err := a.repository.listInvoices(r.Context(), limit, offset, invoiceFilters{
		CustomerID: r.URL.Query().Get("customer_id"),
		OrderID:    r.URL.Query().Get("order_id"),
		Status:     r.URL.Query().Get("status"),
	})
	if err != nil {
		a.metrics.requestsTotal.WithLabelValues("invoices_list", "error").Inc()
		platformHttp.WriteError(w, http.StatusInternalServerError, err.Error(), platformHttp.CorrelationIDFromContext(r.Context()))
		return
	}

	a.metrics.requestsTotal.WithLabelValues("invoices_list", "success").Inc()
	platformHttp.WriteJSON(w, http.StatusOK, result)
}

func (a *app) handleGetInvoice(w http.ResponseWriter, r *http.Request) {
	result, err := a.repository.getInvoice(r.Context(), r.PathValue("invoiceID"))
	if err != nil {
		if notFoundError(err) {
			a.metrics.requestsTotal.WithLabelValues("invoices_get", "not_found").Inc()
			platformHttp.WriteError(w, http.StatusNotFound, "invoice not found", platformHttp.CorrelationIDFromContext(r.Context()))
			return
		}
		a.metrics.requestsTotal.WithLabelValues("invoices_get", "error").Inc()
		platformHttp.WriteError(w, http.StatusInternalServerError, err.Error(), platformHttp.CorrelationIDFromContext(r.Context()))
		return
	}

	a.metrics.requestsTotal.WithLabelValues("invoices_get", "success").Inc()
	platformHttp.WriteJSON(w, http.StatusOK, result)
}

func (a *app) parsePagination(r *http.Request) (int, int, error) {
	limit := a.config.DefaultPageSize
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return 0, 0, fmt.Errorf("limit must be an integer")
		}
		if parsed <= 0 || parsed > a.config.MaxPageSize {
			return 0, 0, fmt.Errorf("limit must be between 1 and %d", a.config.MaxPageSize)
		}
		limit = parsed
	}

	offset := 0
	if raw := r.URL.Query().Get("offset"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return 0, 0, fmt.Errorf("offset must be an integer")
		}
		if parsed < 0 {
			return 0, 0, fmt.Errorf("offset must be zero or positive")
		}
		offset = parsed
	}

	return limit, offset, nil
}
