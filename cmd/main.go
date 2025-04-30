package main

import (
	"ebpf_server/internal/grpc/pb"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedEventCollectorServer
}

func (s *server) SendEvents(stream pb.EventCollector_SendEventsServer) error {
	log.Println("📥 Receiving streamed events...")

	for {
		event, err := stream.Recv()
		if err == io.EOF {
			log.Println("✅ Finished receiving all events.")
			return stream.SendAndClose(&pb.CollectorAck{
				Status:  "OK",
				Message: "All events received successfully",
			})
		}
		if err != nil {
			log.Printf("❌ Error receiving event: %v", err)
			return err
		}

		log.Printf("📡 Event from node %s | type=%s", event.NodeName, event.EventType)
		log.Printf("🔍 Event: PID=%d UID=%d COMM=%s FILENAME=%s RET=%d TS=%d EXIT_TS=%d LAT=%d\n",
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
func main() {
	log.Println("­ƒÜÇ Starting gRPC server on :8080...")

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterEventCollectorServer(grpcServer, &server{})

	log.Println("Ô£à Server ready and listening")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
