package processor

import (
	"context"
	"ebpf_server/internal/grpc/pb"
	"ebpf_server/internal/kafka"
	"ebpf_server/pkg/logutil"
	"ebpf_server/pkg/programs"

	"go.uber.org/zap"
)

type Processor struct{
  kpResource *kafka.KafkaProducer
  kpFrequency *kafka.KafkaProducer
  eventChanResource chan *pb.Batch
  eventChanFreq chan *pb.Batch
  stopCh chan struct{}
}

func NewProcessor (kpResource *kafka.KafkaProducer,kpFrequency *kafka.KafkaProducer) *Processor{

  processor := &Processor{
    kpResource: kpResource,
    kpFrequency: kpFrequency,
    eventChanResource: make(chan *pb.Batch,1000),
    eventChanFreq: make(chan *pb.Batch,1000),
    stopCh: make(chan struct{}),
  }

  go processor.run()

  return processor
}

func (p *Processor) run(){

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

func (p *Processor) Stop() {
	close(p.stopCh)
}

