package kafka

import (
	"context"
	"ebpf_server/internal/grpc/gpupb"
	"ebpf_server/internal/grpc/pb"

	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

type KafkaProducer struct{
  writer *kafka.Writer
  topic string
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer{
  return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic, 
			Balancer: &kafka.LeastBytes{},
		},
		topic: topic,
	}
}

func (kp *KafkaProducer) Send(ctx context.Context,event *pb.Batch) error{

  msgs := make([]kafka.Message, 0, len(event.Batch))

  for _, ev := range event.Batch{
    data,err := proto.Marshal(ev)
    if err !=nil{
      return err
    }

    msgs = append(msgs, kafka.Message{
      Value: data,
    })
  }
  
  
  return kp.writer.WriteMessages(ctx, msgs...)
}

func (kp *KafkaProducer) SendGpu(ctx context.Context,event *gpupb.GpuBatch) error{

  msgs := make([]kafka.Message, 0, len(event.Batch))

  for _, ev := range event.Batch{
    
    switch ev.Payload.(type){
      case *gpupb.GpuEvent_Token:
        continue
      case *gpupb.GpuEvent_Tw:
        data,err := proto.Marshal(ev)
        if err !=nil{
          return err
        }

        msgs = append(msgs, kafka.Message{
          Value: data,
        })
      default:
        continue
    }
    
  }
  
  
  return kp.writer.WriteMessages(ctx, msgs...)
}


func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}
