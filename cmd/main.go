package main

import (
	"context"
	"ebpf_server/internal/clickhouse"
	"ebpf_server/internal/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("Starting gRPC server on :8080...")
  ctx, cancel := context.WithCancel(context.Background())
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
  grpcServer := grpc.NewGrpcServer(server)

  sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

  go func() {
		sig := <-sigCh
		log.Printf("Received signal: %s. Initiating graceful shutdown...", sig)
		grpcServer.GracefulStop()
		cancel()
	}()

	log.Println("Server ready and listening")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
