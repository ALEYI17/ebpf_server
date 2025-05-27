package main

import (
	"context"
	"ebpf_server/internal/clickhouse"
	"ebpf_server/internal/config"
	"ebpf_server/internal/grpc"
	"ebpf_server/pkg/logutil"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {

  logutil.InitLogger()

  logger := logutil.GetLogger()

	ctx, cancel := context.WithCancel(context.Background())
  defer cancel()

  conf := config.LoadServerConfig()

  logger.Info("Starting gRPC server", zap.String("port", conf.Port))
  lis, err := net.Listen("tcp", fmt.Sprintf(":%s", conf.Port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}

  conn,err := clickhouse.NewConnection(ctx,conf)
  
  if err!=nil{
    logger.Fatal("Cannot create the ClickHouse connection", zap.Error(err))
  }

  bi := clickhouse.NewBatchInserter(conn, conf.BatchSize, time.Duration(conf.BatchFlushMs)*time.Millisecond)

  server  := grpc.NewServer(bi)

  grpcServer := grpc.NewGrpcServer(server)

  sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

  go func() {
		sig := <-sigCh
		logger.Info("Received termination signal, initiating graceful shutdown", zap.String("signal", sig.String()))
		grpcServer.Stop()
		cancel()
	}()

	logger.Info("Server is ready and listening")

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", zap.Error(err))
	}
}
