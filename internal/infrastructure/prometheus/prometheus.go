package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusCollector struct {
	httpRequestsTotal      *prometheus.CounterVec
	httpRequestDuration    *prometheus.HistogramVec
	pvzsCreatedTotal       prometheus.Counter
	receptionsCreatedTotal prometheus.Counter
	productsAddedTotal     prometheus.Counter
}

func NewPrometheusCollector() *PrometheusCollector {
	return &PrometheusCollector{
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "code"},
		),
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		),
		pvzsCreatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "pvzs_created_total",
				Help: "Total number of created PVZs",
			},
		),
		receptionsCreatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "receptions_created_total",
				Help: "Total number of created receptions",
			},
		),
		productsAddedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "products_added_total",
				Help: "Total number of added products",
			},
		),
	}
}

func (c *PrometheusCollector) IncPVZsCreated() {
	c.pvzsCreatedTotal.Inc()
}

func (c *PrometheusCollector) IncReceptionsCreated() {
	c.receptionsCreatedTotal.Inc()
}

func (c *PrometheusCollector) IncProductsAdded() {
	c.productsAddedTotal.Inc()
}

func (c *PrometheusCollector) ObserveHTTPRequestDuration(method string, duration float64) {
	c.httpRequestDuration.WithLabelValues(method).Observe(duration)
}

func (c *PrometheusCollector) IncHTTPRequestsTotal(method, code string) {
	c.httpRequestsTotal.WithLabelValues(method, code).Inc()
}
