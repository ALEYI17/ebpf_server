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

		log.Printf(
			"Event from node=%s | type=%s\n"+
				"  PID=%d UID=%d GID=%d PPID=%d User=%s\n"+
				"  UserPID=%d UserPPID=%d\n"+
				"  Comm=%s Filename=%s CgroupName=%s CgroupID=%d\n"+
				"  Ret=%d Latency(ns)=%d StartTS=%d ExitTS=%d",
			event.NodeName,
			event.EventType,
			event.Pid,
			event.Uid,
			event.Gid,
			event.Ppid,
			event.User,
			event.UserPid,
			event.UserPpid,
			event.Comm,
			event.Filename,
			event.CgroupName,
			event.CgroupId,
			event.ReturnCode,
			event.LatencyNs,
			event.TimestampNs,
			event.TimestampNsExit,
		)    
    if err := s.ch.InsertTraceEvent(stream.Context(), event);err!=nil{
      log.Printf("Error inserting data in clickhouse %s", err)
      return err
    }
	}
}

