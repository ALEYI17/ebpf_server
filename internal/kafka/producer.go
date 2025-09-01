package kafka

import (
	"context"
	"ebpf_server/internal/grpc/pb"
  "google.golang.org/protobuf/proto"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct{
  writer *kafka.Writer
  topic string
}

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer{
  return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic, // 👈 fixed topic
			Balancer: &kafka.LeastBytes{},
		},
		topic: topic,
	}
}

func (kp *KafkaProducer) Send(ctx context.Context,event *pb.EbpfEvent) error{

  data,err := proto.Marshal(event)
  if err !=nil{
    return err
  }

  err = kp.writer.WriteMessages(ctx, kafka.Message{
    Value: data,
  })

  return err
}

func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}
