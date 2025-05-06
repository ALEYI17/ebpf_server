package grpc

import (
	"ebpf_server/internal/clickhouse"
	"ebpf_server/internal/grpc/pb"
	"io"
	"log"

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
	log.Println(" Receiving streamed events...")

	for {
		event, err := stream.Recv()
		if err == io.EOF {
			log.Println("Finished receiving all events.")
			return stream.SendAndClose(&pb.CollectorAck{
				Status:  "OK",
				Message: "All events received successfully",
			})
		}
		if err != nil {
			log.Printf("Error receiving event: %v", err)
			return err
		}

		   
    if err := s.ch.InsertTraceEvent(stream.Context(), event);err!=nil{
      log.Printf("Error inserting data in clickhouse %s", err)
      return err
    }
	}
}

