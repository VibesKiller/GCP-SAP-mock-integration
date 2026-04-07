package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gcp-sap-mock-integration/internal/domain"
	"gcp-sap-mock-integration/internal/platform/httpx"
)

type dispatchResult struct {
	Attempted    bool            `json:"attempted"`
	URL          string          `json:"url,omitempty"`
	StatusCode   int             `json:"status_code,omitempty"`
	ResponseBody json.RawMessage `json:"response_body,omitempty"`
}

type simulationResponse struct {
	EventType      string         `json:"event_type"`
	Payload        any            `json:"payload"`
	Dispatched     bool           `json:"dispatched"`
	DispatchResult dispatchResult `json:"dispatch_result,omitempty"`
}

type app struct {
	config     appConfig
	logger     *slog.Logger
	httpClient *http.Client
	samples    sampleCatalog
	metrics    *metrics
}

func newApp(cfg appConfig, logger *slog.Logger, samples sampleCatalog) *app {
	return &app{
		config: cfg,
		logger: logger,
		httpClient: &http.Client{
			Timeout: cfg.DispatchTimeout,
		},
		samples: samples,
		metrics: newMetrics(),
	}
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	httpx.RegisterHealthEndpoints(mux, a.config.ServiceName, nil)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/api/v1/sample-data", a.handleSampleData)
	mux.HandleFunc("/api/v1/simulations/sales-orders/create", a.handleSalesOrderCreate)
	mux.HandleFunc("/api/v1/simulations/sales-orders/update", a.handleSalesOrderUpdate)
	mux.HandleFunc("/api/v1/simulations/customers/update", a.handleCustomerUpdate)
	mux.HandleFunc("/api/v1/simulations/invoices/issue", a.handleInvoiceIssued)

	return httpx.Chain(mux,
		httpx.CorrelationMiddleware(),
		httpx.RecoveryMiddleware(a.logger),
		httpx.LoggingMiddleware(a.logger),
	)
}

func (a *app) handleSampleData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	httpx.WriteJSON(w, http.StatusOK, a.samples)
}

func (a *app) handleSalesOrderCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	payload, err := a.salesOrderPayloadFromRequest(r, a.samples.SalesOrderCreate)
	if err != nil {
		a.metrics.simulationsTotal.WithLabelValues(domain.EventTypeSalesOrderCreated, "invalid").Inc()
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	a.respondWithSimulation(w, r, http.StatusCreated, domain.EventTypeSalesOrderCreated, payload, http.MethodPost, "/api/v1/sap/sales-orders")
}

func (a *app) handleSalesOrderUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	payload, err := a.salesOrderPayloadFromRequest(r, a.samples.SalesOrderUpdate)
	if err != nil {
		a.metrics.simulationsTotal.WithLabelValues(domain.EventTypeSalesOrderUpdated, "invalid").Inc()
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	target := fmt.Sprintf("/api/v1/sap/sales-orders/%s", payload.SalesDocumentID)
	a.respondWithSimulation(w, r, http.StatusOK, domain.EventTypeSalesOrderUpdated, payload, http.MethodPatch, target)
}

func (a *app) handleCustomerUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	payload, err := a.customerPayloadFromRequest(r, a.samples.CustomerUpdate)
	if err != nil {
		a.metrics.simulationsTotal.WithLabelValues(domain.EventTypeCustomerUpdated, "invalid").Inc()
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	target := fmt.Sprintf("/api/v1/sap/customers/%s", payload.CustomerID)
	a.respondWithSimulation(w, r, http.StatusOK, domain.EventTypeCustomerUpdated, payload, http.MethodPatch, target)
}

func (a *app) handleInvoiceIssued(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed", httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	payload, err := a.invoicePayloadFromRequest(r, a.samples.InvoiceIssued)
	if err != nil {
		a.metrics.simulationsTotal.WithLabelValues(domain.EventTypeInvoiceIssued, "invalid").Inc()
		httpx.WriteError(w, http.StatusBadRequest, err.Error(), httpx.CorrelationIDFromContext(r.Context()))
		return
	}

	a.respondWithSimulation(w, r, http.StatusCreated, domain.EventTypeInvoiceIssued, payload, http.MethodPost, "/api/v1/sap/invoices")
}

func (a *app) salesOrderPayloadFromRequest(r *http.Request, fallback domain.SAPSalesOrderPayload) (domain.SAPSalesOrderPayload, error) {
	payload := fallback
	hasBody, err := httpx.DecodeOptionalJSON(r, &payload)
	if err != nil {
		return domain.SAPSalesOrderPayload{}, fmt.Errorf("decode request body: %w", err)
	}

	if !hasBody {
		payload = fallback
	}

	if err := payload.Validate(); err != nil {
		return domain.SAPSalesOrderPayload{}, err
	}

	return payload, nil
}

func (a *app) customerPayloadFromRequest(r *http.Request, fallback domain.SAPCustomerPayload) (domain.SAPCustomerPayload, error) {
	payload := fallback
	hasBody, err := httpx.DecodeOptionalJSON(r, &payload)
	if err != nil {
		return domain.SAPCustomerPayload{}, fmt.Errorf("decode request body: %w", err)
	}

	if !hasBody {
		payload = fallback
	}

	if err := payload.Validate(); err != nil {
		return domain.SAPCustomerPayload{}, err
	}

	return payload, nil
}

func (a *app) invoicePayloadFromRequest(r *http.Request, fallback domain.SAPInvoicePayload) (domain.SAPInvoicePayload, error) {
	payload := fallback
	hasBody, err := httpx.DecodeOptionalJSON(r, &payload)
	if err != nil {
		return domain.SAPInvoicePayload{}, fmt.Errorf("decode request body: %w", err)
	}

	if !hasBody {
		payload = fallback
	}

	if err := payload.Validate(); err != nil {
		return domain.SAPInvoicePayload{}, err
	}

	return payload, nil
}

func (a *app) respondWithSimulation(w http.ResponseWriter, r *http.Request, status int, eventType string, payload any, dispatchMethod, ingestionPath string) {
	response := simulationResponse{
		EventType:  eventType,
		Payload:    payload,
		Dispatched: false,
	}

	shouldDispatch := a.config.AutoDispatch || strings.EqualFold(r.URL.Query().Get("dispatch"), "true")
	if shouldDispatch {
		started := time.Now()
		dispatchResult, err := a.dispatch(r.Context(), ingestionPath, dispatchMethod, payload, httpx.CorrelationIDFromContext(r.Context()))
		a.metrics.dispatchDuration.WithLabelValues(eventType).Observe(time.Since(started).Seconds())
		if err != nil {
			a.metrics.dispatchTotal.WithLabelValues(eventType, "error").Inc()
			a.metrics.simulationsTotal.WithLabelValues(eventType, "dispatch_error").Inc()
			httpx.WriteError(w, http.StatusBadGateway, err.Error(), httpx.CorrelationIDFromContext(r.Context()))
			return
		}
		a.metrics.dispatchTotal.WithLabelValues(eventType, "success").Inc()
		response.Dispatched = true
		response.DispatchResult = dispatchResult
	}

	a.metrics.simulationsTotal.WithLabelValues(eventType, "success").Inc()
	a.logger.Info("sap simulation completed",
		"event_type", eventType,
		"dispatched", response.Dispatched,
		"correlation_id", httpx.CorrelationIDFromContext(r.Context()),
	)
	httpx.WriteJSON(w, status, response)
}

func (a *app) dispatch(ctx context.Context, path, method string, payload any, correlationID string) (dispatchResult, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return dispatchResult{}, fmt.Errorf("marshal dispatch payload: %w", err)
	}

	targetURL := strings.TrimRight(a.config.IngestionBaseURL, "/") + path
	req, err := http.NewRequestWithContext(ctx, method, targetURL, bytes.NewReader(body))
	if err != nil {
		return dispatchResult{}, fmt.Errorf("build dispatch request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if correlationID != "" {
		req.Header.Set("X-Correlation-ID", correlationID)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return dispatchResult{}, fmt.Errorf("dispatch event to ingestion-api: %w", err)
	}
	defer resp.Body.Close()

	payloadBody := json.RawMessage{}
	if content, readErr := io.ReadAll(resp.Body); readErr == nil && len(content) > 0 {
		payloadBody = content
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return dispatchResult{}, fmt.Errorf("ingestion-api returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(payloadBody)))
	}

	return dispatchResult{
		Attempted:    true,
		URL:          targetURL,
		StatusCode:   resp.StatusCode,
		ResponseBody: payloadBody,
	}, nil
}
