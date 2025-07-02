package grpc

import (
	"ebpf_server/internal/clickhouse"
	"ebpf_server/internal/grpc/pb"
	"ebpf_server/internal/metrics"
	"ebpf_server/pkg/logutil"
	"io"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedEventCollectorServer
  bi *clickhouse.BatchInserter
}

func NewServer(bi *clickhouse.BatchInserter) *Server{
  return &Server{bi: bi}
}

func NewGrpcServer(server *Server) *grpc.Server{
  grpcserver := grpc.NewServer()
  
  pb.RegisterEventCollectorServer(grpcserver, server)

  return grpcserver
}

func (s *Server) SendEvents(stream pb.EventCollector_SendEventsServer) error {
  logger := logutil.GetLogger()
  metrics.GrpcActiveStreams.Inc()
  defer metrics.GrpcActiveStreams.Dec()

	logger.Info("Receiving streamed events from client")

	for {
		event, err := stream.Recv()
		if err == io.EOF {
			logger.Info("Finished receiving all events")
			return stream.SendAndClose(&pb.CollectorAck{
				Status:  "OK",
				Message: "All events received successfully",
			})
		}
		if err != nil {
			logger.Error("Error receiving event from stream", zap.Error(err))
      metrics.GrpcStreamErrors.WithLabelValues("recv_error").Inc()
			return err
		}

    metrics.GrpcEventsReceived.WithLabelValues(event.EventType).Inc()
    s.bi.Submit(event)

	}
}

