package metrics

import "github.com/prometheus/client_golang/prometheus"


var (
	ClickhouseInsertTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "infrasight",
			Subsystem: "clickhouse",
			Name:      "insert_batches_total",
			Help:      "Total number of ClickHouse insert batches (successful or failed).",
		},
		[]string{"status"}, // "success", "error"
	)

	ClickhouseInsertLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "infrasight",
			Subsystem: "clickhouse",
			Name:      "insert_latency_seconds",
			Help:      "Histogram of ClickHouse batch insert latency.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{},
	)

	ClickhouseBatchSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "infrasight",
			Subsystem: "clickhouse",
			Name:      "batch_size",
			Help:      "Number of events per ClickHouse insert batch.",
			Buckets:   prometheus.ExponentialBuckets(10, 2, 8), // 10, 20, 40, ..., 1280
		},
		[]string{},
	)
)
