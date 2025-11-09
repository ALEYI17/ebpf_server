package grpc

import (
	"context"
	"ebpf_server/internal/clickhouse"
	"ebpf_server/internal/grpc/gpupb"
	"ebpf_server/internal/grpc/pb"
	"ebpf_server/internal/metrics"
	"ebpf_server/internal/processor"
	"ebpf_server/pkg/logutil"
	"io"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedEventCollectorServer
  gpupb.UnimplementedGpuEventCollectorServer
  bi *clickhouse.BatchInserter
  big *clickhouse.GpuBatchInserter
  p *processor.Processor
}

func NewServer(bi *clickhouse.BatchInserter, big *clickhouse.GpuBatchInserter,p *processor.Processor) *Server{
  return &Server{
    bi: bi,
    big: big,
    p: p,
  }
}

func NewGrpcServer(server *Server) *grpc.Server{
  grpcserver := grpc.NewServer(grpc.MaxRecvMsgSize(64*1024*1024), grpc.MaxSendMsgSize(64*1024*1024))
  
  pb.RegisterEventCollectorServer(grpcserver, server)
  gpupb.RegisterGpuEventCollectorServer(grpcserver, server)

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
    s.p.Submit_event(event)
	}
}


func (s *Server) SendBatch(ctx context.Context,in *pb.Batch) (*pb.CollectorAck,error){
  logger:= logutil.GetLogger()
  logger.Info("Received batch of events", zap.Int("count", len(in.Batch)))
  
  s.p.Submit(in)
  for _, e := range in.Batch{
    s.bi.Submit(e)
  }

  return &pb.CollectorAck{
    Status: "Ok",
    Message: "Batch processed successfully",
  },nil
}

func(s *Server) SendGpuBatch(ctx context.Context, in *gpupb.GpuBatch) (*gpupb.CollectorAck, error){
  logger := logutil.GetLogger()
  logger.Info("Received Gpu batch of events", zap.Int("count", len(in.Batch)))

  for _,e := range in.Batch{
    s.big.Submit(e)
  }
  return &gpupb.CollectorAck{
    Status: "Ok",
    Message: "Batch processed successfully",
  },nil
}
