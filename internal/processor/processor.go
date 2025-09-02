package processor

import (
	"context"
	"ebpf_server/internal/grpc/pb"
	"ebpf_server/internal/kafka"
	"ebpf_server/pkg/logutil"

	"go.uber.org/zap"
)

type Processor struct{
  kpResource *kafka.KafkaProducer
  kpFrequency *kafka.KafkaProducer
  eventChanResource chan *pb.EbpfEvent
  eventChanFreq chan *pb.EbpfEvent
  stopCh chan struct{}
}

func NewProcessor (kpResource *kafka.KafkaProducer,kpFrequency *kafka.KafkaProducer) *Processor{

  processor := &Processor{
    kpResource: kpResource,
    kpFrequency: kpFrequency,
    eventChanResource: make(chan *pb.EbpfEvent,1000),
    eventChanFreq: make(chan *pb.EbpfEvent,1000),
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

func (p *Processor) Submit(event *pb.EbpfEvent){
  
  switch event.Payload.(type){
    case *pb.EbpfEvent_SysFreq:
      select{
        case p.eventChanFreq <- event:
        default:
          logutil.GetLogger().Warn("dropped freq event: channel full", zap.String("comm", event.Comm))
      }
    case *pb.EbpfEvent_Resource:
      select{
        case p.eventChanResource <- event:
        default:
          logutil.GetLogger().Warn("dropped resource event: channel full", zap.String("comm", event.Comm))
      }

    default:
      return
  }
}

func (p *Processor) Stop() {
	close(p.stopCh)
}

