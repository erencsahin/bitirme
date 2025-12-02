package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OrdersCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_created_total",
		Help: "Total number of orders created",
	})

	OrdersFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_failed_total",
		Help: "Total number of failed orders",
	})

	OrdersCancelled = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_cancelled_total",
		Help: "Total number of cancelled orders",
	})

	OrderDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "order_creation_duration_seconds",
		Help:    "Duration of order creation in seconds",
		Buckets: prometheus.DefBuckets,
	})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latencies in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "endpoint", "status"})

	StockCheckDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "stock_check_duration_seconds",
		Help:    "Duration of stock check calls",
		Buckets: prometheus.DefBuckets,
	})

	PaymentProcessDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "payment_process_duration_seconds",
		Help:    "Duration of payment processing",
		Buckets: prometheus.DefBuckets,
	})
)
