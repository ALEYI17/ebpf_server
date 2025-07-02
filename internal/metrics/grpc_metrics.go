package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	GrpcEventsReceived = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "infrasight",
			Subsystem: "grpc",
			Name:      "events_received_total",
			Help:      "Total number of eBPF events received via gRPC stream.",
		},
		[]string{"source"}, // Optional label if you can extract event type
	)

	GrpcStreamErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "infrasight",
			Subsystem: "grpc",
			Name:      "stream_errors_total",
			Help:      "Total number of gRPC stream receive errors.",
		},
		[]string{"type"},
	)

	GrpcActiveStreams = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "infrasight",
			Subsystem: "grpc",
			Name:      "active_streams",
			Help:      "Number of active gRPC streams.",
		},
	)
)
