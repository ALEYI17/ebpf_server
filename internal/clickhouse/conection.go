package clickhouse

import (
	"context"
	"ebpf_server/internal/grpc/pb"
	"fmt"
	"strings"
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

func (ch *Chconnection) InsertTraceEvent(ctx context.Context,event *pb.EbpfEvent) error{
  
  query := fmt.Sprintf(
		`INSERT INTO audit.tracing_events 
		(pid, uid, gid, ppid, user, user_pid, user_ppid, comm, filename, cgroup_name, cgroup_id,
		 timestamp_ns_exit, return_code, timestamp_ns, latency_ns, event_type, node_name) 
		VALUES (%d, %d, %d, %d, '%s', %d, %d, '%s', '%s', '%s', %d, %d, %d, %d, %d, '%s', '%s')`,
		event.Pid,
		event.Uid,
		event.Gid,
		event.Ppid,
		escapeSQLString(event.User),
		event.UserPid,
		event.UserPpid,
		escapeSQLString(event.Comm),
		escapeSQLString(event.Filename),
		escapeSQLString(event.CgroupName),
		event.CgroupId,
		event.TimestampNsExit,
		event.ReturnCode,
		event.TimestampNs,
		event.LatencyNs,
		escapeSQLString(event.EventType),
		escapeSQLString(event.NodeName),
	) 

  fmt.Print(query)
  if err := ch.conn.AsyncInsert(ctx, query, false); err !=nil{
    return err
  }
  return nil
}

func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''") // escape single quotes
}
