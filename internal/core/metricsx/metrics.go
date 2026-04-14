package metricsx

import (
	"context"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	once     sync.Once
	registry *prometheus.Registry

	startedAt = time.Now()

	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_http_requests_total",
			Help: "Total HTTP requests handled by auth-service.",
		},
		[]string{"service", "method", "route", "status_class"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "route", "status_class"},
	)

	httpErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_http_errors_total",
			Help: "Total HTTP requests with status >= 400.",
		},
		[]string{"service", "method", "route", "status_class"},
	)
)

func setupRegistry(serviceName string) *prometheus.Registry {
	once.Do(func() {
		registry = prometheus.NewRegistry()
		registry.MustRegister(collectors.NewGoCollector())
		registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
		registry.MustRegister(httpRequests, httpRequestDuration, httpErrors)

		registry.MustRegister(prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "auth_app_uptime_seconds",
				Help: "Application uptime in seconds.",
				ConstLabels: prometheus.Labels{
					"service": serviceName,
				},
			},
			func() float64 {
				return time.Since(startedAt).Seconds()
			},
		))
	})
	return registry
}

func ObserveHTTPRequest(serviceName, method, route string, statusCode int, duration time.Duration) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = "UNKNOWN"
	}
	if strings.TrimSpace(route) == "" {
		route = "unknown"
	}
	statusClass := classifyStatus(statusCode)

	labels := prometheus.Labels{
		"service":      serviceName,
		"method":       method,
		"route":        route,
		"status_class": statusClass,
	}

	httpRequests.With(labels).Inc()
	httpRequestDuration.With(labels).Observe(duration.Seconds())
	if statusCode >= 400 {
		httpErrors.With(labels).Inc()
	}
}

func StartPushLoop(serviceName, pushGatewayURL, job, instance string, interval time.Duration) func() {
	if !isValidPushGatewayURL(pushGatewayURL) {
		slog.Warn("metrics push disabled: invalid PUSHGATEWAY_URL", "url", pushGatewayURL)
		return func() {}
	}
	if strings.TrimSpace(job) == "" {
		job = serviceName
	}
	if strings.TrimSpace(instance) == "" {
		instance = serviceName
	}
	if interval <= 0 {
		interval = 15 * time.Second
	}

	registry := setupRegistry(serviceName)
	ctx, cancel := context.WithCancel(context.Background())

	pushOnce := func() {
		err := push.New(pushGatewayURL, job).
			Gatherer(registry).
			Grouping("instance", instance).
			Push()
		if err != nil {
			slog.Warn("push metrics to Pushgateway failed", "log_type", "metrics", "error", err.Error())
		}
	}

	pushOnce()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pushOnce()
			}
		}
	}()

	return cancel
}

func classifyStatus(statusCode int) string {
	switch {
	case statusCode >= 500:
		return "5xx"
	case statusCode >= 400:
		return "4xx"
	case statusCode >= 300:
		return "3xx"
	case statusCode >= 200:
		return "2xx"
	default:
		return "1xx"
	}
}

func isValidPushGatewayURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	if u.Scheme == "" || u.Host == "" {
		return false
	}
	return true
}
