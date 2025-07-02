package metrics

import "github.com/prometheus/client_golang/prometheus"

func RegisterAll() {
	prometheus.MustRegister(
		GrpcEventsReceived,
		GrpcStreamErrors,
		GrpcActiveStreams,
		ClickhouseInsertTotal,
		ClickhouseInsertLatency,
		ClickhouseBatchSize,
	)
}
