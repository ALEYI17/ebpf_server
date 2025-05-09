package grpc

import (
	"ebpf_server/internal/clickhouse"
	"ebpf_server/internal/grpc/pb"
	"ebpf_server/pkg/logutil"
	"io"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedEventCollectorServer
  ch *clickhouse.Chconnection
}

func NewServer(conn *clickhouse.Chconnection) *Server{
  return &Server{ch: conn}
}

func NewGrpcServer(server *Server) *grpc.Server{
  grpcserver := grpc.NewServer()
  
  pb.RegisterEventCollectorServer(grpcserver, server)

  return grpcserver
}

func (s *Server) SendEvents(stream pb.EventCollector_SendEventsServer) error {
  logger := logutil.GetLogger()
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
			return err
		}

		   
    if err := s.ch.InsertTraceEvent(stream.Context(), event);err!=nil{
      logger.Error("Failed to insert event into ClickHouse", zap.Error(err))
      return err
    }
	}
}

