package main

import (
	"context"
	"ebpf_server/internal/clickhouse"
	"ebpf_server/internal/grpc"
	"log"
	"net"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("­ƒÜÇ Starting gRPC server on :8080...")
  ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
  defer cancel()

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

  conn,err := clickhouse.NewConnection(ctx)
  if err!=nil{
    log.Fatalf("Cannor create the clickhouse connection %s", err)
  }
  server  := grpc.NewServer(conn)
  grpcServer := grpc.RegisterGrpcServer(server)

	log.Println("Ô£à Server ready and listening")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
