package clickhouse

import (
	"context"
	"crypto/tls"
	"ebpf_server/internal/grpc/pb"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Chconnection struct {
	conn driver.Conn
}

func NewConnection(ctx context.Context) (*Chconnection, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"localhost:9000"},
		Auth: clickhouse.Auth{
			Database: "audit",
			Username: "user",
			Password: "password",
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "ebpf_server", Version: "0.1"},
			},
		},
    Debug: true,
		Debugf: func(format string, v ...interface{}) {
			fmt.Printf(format, v)
		},
    DialTimeout:  time.Duration(10) * time.Second,
    MaxOpenConns:     5,
    MaxIdleConns:     5,
    BlockBufferSize: 10,
    Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		TLS: &tls.Config{
			InsecureSkipVerify: true,
		},
	})

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}

	return &Chconnection{conn: conn}, nil
}

func (ch *Chconnection) InsertTraceEvent(ctx context.Context,event *pb.EbpfEvent_ExecveEvent) error{
  
  query := fmt.Sprintf("INSERT INTO audit.tracing_events VALUES(%d,%d,%s,%s,%d,%d,%d,%d)",event.ExecveEvent.Pid,event.ExecveEvent.Uid,
    event.ExecveEvent.Comm , event.ExecveEvent.Filename,event.ExecveEvent.TimestampNsExit,event.ExecveEvent.ReturnCode,event.ExecveEvent.TimestampNs,
    event.ExecveEvent.LatencyNs)

  if err := ch.conn.AsyncInsert(ctx, query, false); err !=nil{
    return err
  }
  return nil
}
