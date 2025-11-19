package processor

import (
	"context"
	"ebpf_server/internal/grpc/gpupb"
	"ebpf_server/internal/grpc/pb"
	"ebpf_server/internal/kafka"
	"ebpf_server/pkg/logutil"
	"ebpf_server/pkg/programs"
	"time"

	"go.uber.org/zap"
)

type Processor struct{
  kpResource *kafka.KafkaProducer
  kpFrequency *kafka.KafkaProducer
  kpStatic *kafka.KafkaProducer
  kpGpu   *kafka.KafkaProducer
  eventChanResource chan *pb.Batch
  eventChanFreq chan *pb.Batch
  eventChanStatic chan *pb.EbpfEvent
  eventChanGpu chan *gpupb.GpuBatch
  batchSize     int
  flushInterval time.Duration
  stopCh chan struct{}
}

func NewProcessor (kpResource *kafka.KafkaProducer,kpFrequency *kafka.KafkaProducer,kpStatic *kafka.KafkaProducer,kpgpu *kafka.KafkaProducer ,batchSize int , flushInterval time.Duration) *Processor{

  processor := &Processor{
    kpResource: kpResource,
    kpFrequency: kpFrequency,
    kpStatic: kpStatic,
    kpGpu: kpgpu,
    eventChanResource: make(chan *pb.Batch,1000),
    eventChanFreq: make(chan *pb.Batch,1000),
    eventChanStatic: make(chan *pb.EbpfEvent,1000),
    eventChanGpu: make(chan *gpupb.GpuBatch,1000),
    batchSize: batchSize,
    flushInterval: flushInterval,
    stopCh: make(chan struct{}),
  }

  go processor.run()

  return processor
}

func (p *Processor) run(){

  ticker := time.NewTicker(p.flushInterval)
  defer ticker.Stop()
  var staticBuffer []*pb.EbpfEvent

  for{

    select{
    case <- p.stopCh:
      logutil.GetLogger().Info("stoping TODO")
      return

    case ever := <- p.eventChanResource:
      err := p.kpResource.Send(context.Background(), ever)
      if err !=nil{
        logutil.GetLogger().Error("failed to write to Kafka", zap.Error(err))
      }
    case evef := <- p.eventChanFreq:
      err := p.kpFrequency.Send(context.Background(), evef)
      if err !=nil {
        logutil.GetLogger().Error("failed to write to Kafka", zap.Error(err))
      }
    case eves := <- p.eventChanStatic:
      staticBuffer = append(staticBuffer, eves)
      if len(staticBuffer) >= p.batchSize{
        staticBuffer = p.flushStatic(staticBuffer)
      }
      
    case eveg := <- p.eventChanGpu:
      err := p.kpGpu.SendGpu(context.Background(), eveg)
      if err !=nil {
        logutil.GetLogger().Error("failed to write to Kafka", zap.Error(err))
      }
    case <- ticker.C:
      staticBuffer = p.flushStatic(staticBuffer)
    
    }
  }
}

func (p *Processor) Submit(event *pb.Batch){
  
  switch event.Type{
    case programs.LoadSyscallFreq:
      select{
        case p.eventChanFreq <- event:
        default:
          logutil.GetLogger().Warn("dropped freq event: channel full" )
      }
    case programs.LoadResource:
      select{
        case p.eventChanResource <- event:
        default:
          logutil.GetLogger().Warn("dropped resource event: channel full")
      }

    default:
      return
  }
}

func (p *Processor) Submit_event(event *pb.EbpfEvent){
  
  select{
  case p.eventChanStatic <- event:
  default:
    logutil.GetLogger().Warn("dropped static analysis event: chan full")
  }
}

func (p *Processor) Submit_gpu(batch *gpupb.GpuBatch){

  select{
  case p.eventChanGpu <- batch:
  default:
    logutil.GetLogger().Warn("dropped gpu event: chan full")
  }
}

func (p *Processor) flushStatic(staticBuffer []*pb.EbpfEvent) []*pb.EbpfEvent {
    if len(staticBuffer) == 0 {
        return staticBuffer
    }
    b := pb.Batch{
        Type:  "static",
        Batch: staticBuffer,
    }
    if err := p.kpStatic.Send(context.Background(), &b); err != nil {
        logutil.GetLogger().Error("failed to write static batch to Kafka", zap.Error(err))
    }
    return nil // reset
}

func (p *Processor) Stop() {
	close(p.stopCh)
}

