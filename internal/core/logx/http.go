package logx

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"auth-service/internal/core/metricsx"

	"github.com/gorilla/mux"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

var metricsServiceName = resolveServiceName()

const maxLoggedPayloadBytes = 4096

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	if sr.status == 0 {
		sr.status = http.StatusOK
	}
	n, err := sr.ResponseWriter.Write(b)
	sr.size += n
	return n, err
}

func RequestMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			payload := readRequestPayload(r)
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(recorder, r)

			status := recorder.status
			if status == 0 {
				status = http.StatusOK
			}
			duration := time.Since(start)

			attrs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"size", recorder.size,
				"remote_addr", strings.Split(r.RemoteAddr, ":")[0],
				"user_agent", r.UserAgent(),
				"payload", string(payload),
			}

			route := r.URL.Path
			if current := mux.CurrentRoute(r); current != nil {
				if template, err := current.GetPathTemplate(); err == nil && template != "" {
					route = template
				}
			}
			metricsx.ObserveHTTPRequest(metricsServiceName, r.Method, route, status, duration)

			switch {
			case status >= 500:
				slog.Error("http request completed", append(attrs, "log_type", "http_error")...)
			case status >= 400:
				slog.Warn("http request completed", append(attrs, "log_type", "http_error")...)
			default:
				slog.Info("http request completed", append(attrs, "log_type", "http_access")...)
			}
		})
	}
}

func readRequestPayload(r *http.Request) string {
	if r == nil || r.Body == nil {
		return ""
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("failed to read request body", "log_type", "http_error", "error", err)
		return ""
	}

	r.Body = io.NopCloser(bytes.NewBuffer(body))
	if len(body) == 0 {
		return ""
	}

	if len(body) > maxLoggedPayloadBytes {
		return string(body[:maxLoggedPayloadBytes]) + "...(truncated)"
	}

	return string(body)
}

func resolveServiceName() string {
	v := strings.TrimSpace(os.Getenv("SERVICE_NAME"))
	if v == "" {
		return "auth-service"
	}
	return v
}

func RecoverMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					slog.Error("panic recovered",
						"log_type", "panic",
						"method", r.Method,
						"path", r.URL.Path,
						"panic", fmt.Sprint(recovered),
					)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
