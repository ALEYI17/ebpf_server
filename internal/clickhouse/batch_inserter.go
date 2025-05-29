package clickhouse

import (
	"context"
	"ebpf_server/internal/grpc/pb"
	"ebpf_server/pkg/logutil"
	"time"

	"go.uber.org/zap"
)

type BatchInserter struct{
  ch *Chconnection
  eventChan chan *pb.EbpfEvent
  batchSize int
  flushInterval time.Duration
  stopCh chan struct{}
}

func NewBatchInserter(ch * Chconnection,batchSize int,flushInterval time.Duration) * BatchInserter{
  inserter := &BatchInserter{
    ch: ch,
    eventChan: make(chan *pb.EbpfEvent,500),
    batchSize: batchSize,
    flushInterval: flushInterval,
    stopCh: make(chan struct{}),
  }

  go inserter.run()
  return inserter
}

func (b *BatchInserter) run(){
  ticker := time.NewTicker(b.flushInterval)
  defer ticker.Stop()
  
  var buffer []*pb.EbpfEvent
  for {
    select{
      case <- b.stopCh:
        if len(buffer) > 0 {
          b.ch.InsertBatchTraceEvent(context.Background(), buffer)
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

func (b *BatchInserter) Stop() {
	close(b.stopCh)
}

func (b *BatchInserter) Submit(event *pb.EbpfEvent) {
	select {
	case b.eventChan <- event:
	default:
		// Optional: drop or log if channel is full
		logutil.GetLogger().Warn("Event channel full, dropping event")
	}
}

func (b *BatchInserter) sendBatch(events []*pb.EbpfEvent) {
  var snoopBatch []*pb.EbpfEvent
  var networkBatch []*pb.EbpfEvent

  for _,e := range events{
    switch e.Payload.(type){
      case *pb.EbpfEvent_Snoop:
        snoopBatch = append(snoopBatch, e)
      case *pb.EbpfEvent_Network:
        networkBatch =append(networkBatch, e)
      default:
        logutil.GetLogger().Warn("Unknown event payload type", zap.String("event_type", e.EventType))
    }
  }

  if len(snoopBatch)>0{
    err := b.ch.insertSnoopEvent(context.Background(), snoopBatch)
    if err != nil {
      logutil.GetLogger().Error("Failed to insert batch", zap.Error(err))
    }
  }

  if len(networkBatch) > 0{
    err := b.ch.insertNetworkEvent(context.Background(), networkBatch)
    if err !=nil{
      logutil.GetLogger().Error("Failed to insert batch", zap.Error(err))
    }
  }
}
