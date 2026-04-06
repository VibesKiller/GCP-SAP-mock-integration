package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const correlationIDKey contextKey = "correlation_id"

type ErrorResponse struct {
	Error         string `json:"error"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	wrapped := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}
	return wrapped
}

func CorrelationMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			correlationID := r.Header.Get("X-Correlation-ID")
			if correlationID == "" {
				correlationID = uuid.NewString()
			}

			ctx := context.WithValue(r.Context(), correlationIDKey, correlationID)
			w.Header().Set("X-Correlation-ID", correlationID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(recorder, r)
			logger.Info("http request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.status,
				"duration_ms", time.Since(started).Milliseconds(),
				"correlation_id", CorrelationIDFromContext(r.Context()),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

func RecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("http panic recovered",
						"panic", recovered,
						"path", r.URL.Path,
						"correlation_id", CorrelationIDFromContext(r.Context()),
						"stack", string(debug.Stack()),
					)
					WriteError(w, http.StatusInternalServerError, "internal server error", CorrelationIDFromContext(r.Context()))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func CorrelationIDFromContext(ctx context.Context) string {
	if value, ok := ctx.Value(correlationIDKey).(string); ok {
		return value
	}
	return ""
}

func DecodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(io.LimitReader(r.Body, 2<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}

	if decoder.More() {
		return io.ErrUnexpectedEOF
	}

	return nil
}

func DecodeOptionalJSON(r *http.Request, dst any) (bool, error) {
	defer r.Body.Close()

	body, err := io.ReadAll(io.LimitReader(r.Body, 2<<20))
	if err != nil {
		return false, err
	}

	if len(bytes.TrimSpace(body)) == 0 {
		return false, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return false, err
	}

	return true, nil
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}

	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, message, correlationID string) {
	WriteJSON(w, status, ErrorResponse{
		Error:         message,
		CorrelationID: correlationID,
	})
}

func RegisterHealthEndpoints(mux *http.ServeMux, service string, ready func(context.Context) error) {
	liveHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, HealthResponse{Status: "ok", Service: service})
	})

	readyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ready != nil {
			if err := ready(r.Context()); err != nil {
				WriteJSON(w, http.StatusServiceUnavailable, HealthResponse{Status: "degraded", Service: service})
				return
			}
		}
		WriteJSON(w, http.StatusOK, HealthResponse{Status: "ok", Service: service})
	})

	mux.Handle("/live", liveHandler)
	mux.Handle("/health", readyHandler)
	mux.Handle("/ready", readyHandler)
}
