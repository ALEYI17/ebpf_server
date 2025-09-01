package processor

import (
	"ebpf_server/internal/config"
	"ebpf_server/internal/grpc/pb"
	"ebpf_server/internal/kafka"
	"ebpf_server/pkg/logutil"

	"go.uber.org/zap"
)

type Processor struct{
  kp *kafka.KafkaProducer
  conf *config.ServerConfig
  eventChanResource chan *pb.EbpfEvent
  eventChanFreq chan *pb.EbpfEvent
  stopCh chan struct{}
}

func NewProcessor (kp *kafka.KafkaProducer, conf *config.ServerConfig) *Processor{

  processor := &Processor{
    kp: kp,
    conf: conf,
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
      logutil.GetLogger().Info("doing something for resource payload", zap.String("comm", ever.Comm))
    case evef := <- p.eventChanFreq:
      logutil.GetLogger().Info("doing something for frequency payload", zap.String("comm", evef.Comm))
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

