package clickhouse

import (
	"context"
	"ebpf_server/internal/grpc/gpupb"
	"ebpf_server/internal/metrics"
	"ebpf_server/pkg/logutil"
	"time"

	"go.uber.org/zap"
)

type GpuBatchInserter struct{
  ch *Chconnection
  eventChan chan *gpupb.GpuEvent
  batchSize int
  flushInterval time.Duration
  stopCh chan struct{}
}

func NewgPUBatchInserter(ch * Chconnection,batchSize int,flushInterval time.Duration) * GpuBatchInserter{
  inserter := &GpuBatchInserter{
    ch: ch,
    eventChan: make(chan *gpupb.GpuEvent,500),
    batchSize: batchSize,
    flushInterval: flushInterval,
    stopCh: make(chan struct{}),
  }

  go inserter.run()
  return inserter
}

func (b *GpuBatchInserter) run(){
  ticker := time.NewTicker(b.flushInterval)
  defer ticker.Stop()
  
  var buffer []*gpupb.GpuEvent
  for {
    select{
      case <- b.stopCh:
        if len(buffer) > 0 {
          b.ch.InsertBatchGpuEvent(context.Background(), buffer)
        }
        return

      case <- ticker.C:
        if len(buffer) > 0{
          toSend := buffer
          buffer = nil
          go b.sendBatch(toSend)
        }
      case eve := <- b.eventChan:
        buffer = append(buffer, eve)
        if len(buffer) >= b.batchSize{
          toSend := buffer
          buffer = nil
          go b.sendBatch(toSend)
        }
    }
  }
}

func (b *GpuBatchInserter) Stop() {
	close(b.stopCh)
}

func (b *GpuBatchInserter) Submit(event *gpupb.GpuEvent) {
	select {
	case b.eventChan <- event:
	default:
		// Optional: drop or log if channel is full
		logutil.GetLogger().Warn("Event channel full, dropping event")
	}
}

func (b *GpuBatchInserter) sendBatch(events []*gpupb.GpuEvent){
  var tokenBatch []*gpupb.GpuEvent
  var twBatch []*gpupb.GpuEvent

  for _,e := range events{
    switch e.Payload.(type){
      case *gpupb.GpuEvent_Token:
        tokenBatch = append(tokenBatch, e)
      case *gpupb.GpuEvent_Tw:
        twBatch = append(twBatch, e)
      default:
        logutil.GetLogger().Warn("Unknown event payload type", zap.String("event_type", e.EventType))
    }
  }

  if len(tokenBatch) >0 {
    start:= time.Now()
    err := b.ch.insertGpuTokenEvents(context.Background(), tokenBatch)
    duration := time.Since(start).Seconds()
    metrics.ClickhouseInsertLatency.WithLabelValues().Observe(duration)
    if err != nil {
      metrics.ClickhouseInsertTotal.WithLabelValues("error").Inc()
      logutil.GetLogger().Error("Failed to insert batch", zap.Error(err))
    }else{
      metrics.ClickhouseInsertTotal.WithLabelValues("success").Inc()
    }
  }

  if len(twBatch)>0{
    start:= time.Now()
    err := b.ch.insertGpuTimeWindowEvents(context.Background(), twBatch)
    duration := time.Since(start).Seconds()
    metrics.ClickhouseInsertLatency.WithLabelValues().Observe(duration)
    if err != nil {
      metrics.ClickhouseInsertTotal.WithLabelValues("error").Inc()
      logutil.GetLogger().Error("Failed to insert batch", zap.Error(err))
    }else{
      metrics.ClickhouseInsertTotal.WithLabelValues("success").Inc()
    }
  }
}

