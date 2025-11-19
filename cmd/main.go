package main

import (
	"context"
	"ebpf_server/internal/clickhouse"
	"ebpf_server/internal/config"
	"ebpf_server/internal/grpc"
	"ebpf_server/internal/kafka"
	"ebpf_server/internal/metrics"
	"ebpf_server/internal/processor"
	"ebpf_server/pkg/logutil"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {

  logutil.InitLogger()

  logger := logutil.GetLogger()
  
  conf := config.LoadServerConfig()

  go func(){
    metrics.RegisterAll()
    mux:= http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())
    logger := logutil.GetLogger()
    addr := ":" + conf.PrometheusPort
    logger.Info("Serving Prometheus metrics on port", zap.String("prometheus-port", conf.PrometheusPort))
    if err := http.ListenAndServe(addr, mux); err != nil {
        logger.Warn("Prometheus metrics cannot be served", zap.Error(err))
    }
  }()
	ctx, cancel := context.WithCancel(context.Background())
  defer cancel()

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
  big := clickhouse.NewgPUBatchInserter(conn, conf.BatchSize,time.Duration(conf.BatchFlushMs)*time.Millisecond )

  kpResource := kafka.NewKafkaProducer(conf.KafkaBrokers, "resource")

  kpFrequency := kafka.NewKafkaProducer(conf.KafkaBrokers, "frequency")

  kpStatic := kafka.NewKafkaProducer(conf.KafkaBrokers, "ebpf_events")

  kpgpu:= kafka.NewKafkaProducer(conf.KafkaBrokers, "gpu_fingerprint")

  p := processor.NewProcessor(kpResource , kpFrequency , kpStatic,kpgpu, conf.BatchSize,time.Duration(conf.BatchFlushMs)*time.Millisecond)

  server  := grpc.NewServer(bi,big,p)

  grpcServer := grpc.NewGrpcServer(server)

  sigCh := make(chan os.Signal, 1)

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

  go func() {
		sig := <-sigCh
		logger.Info("Received termination signal, initiating graceful shutdown", zap.String("signal", sig.String()))
		grpcServer.Stop()
    p.Stop()
    kpResource.Close()
    kpFrequency.Close()
    bi.Stop()
		cancel()
	}()

	logger.Info("Server is ready and listening")

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", zap.Error(err))
	}
}
