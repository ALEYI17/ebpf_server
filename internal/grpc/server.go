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

func RegisterGrpcServer(server *Server) *grpc.Server{
  grpcserver := grpc.NewServer()
  
  pb.RegisterEventCollectorServer(grpcserver, server)

  return grpcserver
}

func (s *Server) SendEvents(stream pb.EventCollector_SendEventsServer) error {
	log.Println("­ƒôÑ Receiving streamed events...")

	for {
		event, err := stream.Recv()
		if err == io.EOF {
			log.Println("Ô£à Finished receiving all events.")
			return stream.SendAndClose(&pb.CollectorAck{
				Status:  "OK",
				Message: "All events received successfully",
			})
		}
		if err != nil {
			log.Printf("ÔØî Error receiving event: %v", err)
			return err
		}

		log.Printf("­ƒôí Event from node %s | type=%s", event.NodeName, event.EventType)
		log.Printf("­ƒöì Event: PID=%d UID=%d COMM=%s FILENAME=%s RET=%d TS=%d EXIT_TS=%d LAT=%d\n",
			event.Pid,
			event.Uid,
			event.Comm,
			event.Filename,
			event.ReturnCode,
			event.TimestampNs,
			event.TimestampNsExit,
			event.LatencyNs,
		)
	}
}

