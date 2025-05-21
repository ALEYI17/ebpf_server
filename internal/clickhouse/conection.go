package clickhouse

import (
	"context"
	"ebpf_server/internal/config"
	"ebpf_server/internal/grpc/pb"
	"ebpf_server/pkg/logutil"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.uber.org/zap"
)

type Chconnection struct {
	conn driver.Conn
}

func NewConnection(ctx context.Context, conf *config.ServerConfig) (*Chconnection, error) {
  logger := logutil.GetLogger()
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{conf.DBAddress},
		Auth: clickhouse.Auth{
			Database: conf.DBName,
			Username: conf.DBUser,
			Password: conf.DBPassword,
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
    logger.Error("Failed to open ClickHouse connection", zap.Error(err))
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			logger.Error("ClickHouse exception during ping",
				zap.Int32("code", exception.Code),
				zap.String("message", exception.Message),
				zap.String("stacktrace", exception.StackTrace),
			)
		}
		return nil, err
	}
  logger.Info("Successfully connected to ClickHouse")
	return &Chconnection{conn: conn}, nil
}

func (ch *Chconnection) InsertTraceEvent(ctx context.Context,event *pb.EbpfEvent) error{
  
  logger := logutil.GetLogger()
  query := fmt.Sprintf(
	`INSERT INTO audit.tracing_events 
	(pid, uid, gid, ppid, user, user_pid, user_ppid, comm, filename, cgroup_name, cgroup_id,
	 monotonic_ts_exit_ns, return_code, monotonic_ts_enter_ns, latency_ns, event_type, node_name, latency_ms,wall_time_ms, wall_time_dt,container_id,
   container_image ) 
	VALUES (
		%d, %d, %d, %d, '%s', %d, %d, '%s', '%s', '%s', %d,
		%d, %d, %d, %d, '%s', '%s', %f,%d,fromUnixTimestamp(%d),'%s','%s' )`,
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
	float64(event.LatencyNs)/1_000_000.0,
  event.TimestampUnixMs,
  event.TimestampUnixMs / 1000,
  escapeSQLString(event.ContainerId),
  escapeSQLString(event.ContainerImage),
  )
  
  logger.Debug("Executing ClickHouse insert query", zap.String("query", query))
  
  if err := ch.conn.AsyncInsert(ctx, query, false); err !=nil{
    return err
  }
  return nil
}

func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''") // escape single quotes
}
