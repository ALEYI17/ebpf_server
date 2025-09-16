package clickhouse

import (
	"context"
	"ebpf_server/internal/config"
	"ebpf_server/internal/grpc/pb"
	"ebpf_server/pkg/logutil"
	"encoding/json"
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

// func (ch *Chconnection) InsertTraceEvent(ctx context.Context,event *pb.EbpfEvent) error{
//   logger := logutil.GetLogger()  
//   labelsJSON, err := json.Marshal(event.ContainerLabelsJson)
//   if err != nil {
//     logger.Warn("Could not marshal container labels", zap.Error(err))
//     labelsJSON = []byte("{}") // fallback to empty object
//   }
//   query := fmt.Sprintf(
// 	`INSERT INTO audit.tracing_events 
// 	(pid, uid, gid, ppid, user, user_pid, user_ppid, comm, filename, cgroup_name, cgroup_id,
// 	 monotonic_ts_exit_ns, return_code, monotonic_ts_enter_ns, latency_ns, event_type, node_name, latency_ms,wall_time_ms, wall_time_dt,container_id,
//    container_image ,container_labels_json) 
// 	VALUES (
// 		%d, %d, %d, %d, '%s', %d, %d, '%s', '%s', '%s', %d,
// 		%d, %d, %d, %d, '%s', '%s', %f,%d,toDateTime64(%f, 3),'%s','%s', '%s' )`,
// 	event.Pid,
// 	event.Uid,
// 	event.Gid,
// 	event.Ppid,
// 	escapeSQLString(event.User),
// 	event.UserPid,
// 	event.UserPpid,
// 	escapeSQLString(event.Comm),
// 	escapeSQLString(event.Filename),
// 	escapeSQLString(event.CgroupName),
// 	event.CgroupId,
// 	event.TimestampNsExit,
// 	event.ReturnCode,
// 	event.TimestampNs,
// 	event.LatencyNs,
// 	escapeSQLString(event.EventType),
// 	escapeSQLString(event.NodeName),
// 	float64(event.LatencyNs)/1_000_000.0,
//   event.TimestampUnixMs,
//   float64(event.TimestampUnixMs) / 1000,
//   escapeSQLString(event.ContainerId),
//   escapeSQLString(event.ContainerImage),
//   escapeSQLString(string(labelsJSON)),
//   )
//
//   logger.Debug("Executing ClickHouse insert query", zap.String("query", query))
//
//   if err := ch.conn.AsyncInsert(ctx, query, false); err !=nil{
//     return err
//   }
//   return nil
// }

func (ch *Chconnection) InsertBatchTraceEvent(ctx context.Context,events []*pb.EbpfEvent) []error{

  var snoopEvents []*pb.EbpfEvent
  var networkEvent []*pb.EbpfEvent
  var errors []error
  for _,e := range events{
    if isNetworkevent(e.EventType){
      networkEvent = append(networkEvent, e)
    }else if isSnoopevent(e.EventType){
      snoopEvents = append(snoopEvents, e)
    }
  }

  if len(snoopEvents)>0 {
    if err := ch.insertSnoopEvent(ctx, snoopEvents); err !=nil{
      errors = append(errors, err)
    }
  }

  if len(networkEvent) > 0 {
    if err := ch.insertNetworkEvent(ctx, snoopEvents); err !=nil{
      errors = append(errors, err)
    }
  }

  if len(errors) >0{
    return errors
  }

  return nil
}

func (ch *Chconnection) insertSnoopEvent(ctx context.Context,events []*pb.EbpfEvent) error{
  logger := logutil.GetLogger()

	batch, err := ch.conn.PrepareBatch(ctx, `
		INSERT INTO audit.tracing_events (
			pid, uid, gid, ppid, user, user_pid, user_ppid, comm, filename, cgroup_name, cgroup_id,
			monotonic_ts_exit_ns, return_code, monotonic_ts_enter_ns, latency_ns, event_type, node_name,
			latency_ms, wall_time_ms, wall_time_dt, container_id, container_image, container_labels_json
		)
	`)
	if err != nil {
		logger.Error("Failed to prepare ClickHouse batch", zap.Error(err))
		return err
	}

	for _, event := range events {
    snoopPayload, ok := event.Payload.(*pb.EbpfEvent_Snoop)
		if !ok {
			logger.Warn("Skipping event: unexpected payload type", zap.String("event_type", event.EventType))
			continue
		}
		snoop := snoopPayload.Snoop
		labelsJSON, err := json.Marshal(event.ContainerLabelsJson)
		if err != nil {
			logger.Warn("Could not marshal container labels", zap.Error(err))
			labelsJSON = []byte("{}")
		}

    wallTime := time.Unix(0, int64(event.TimestampUnixMs)*int64(time.Millisecond))
		err = batch.Append(
			event.Pid,
			event.Uid,
			event.Gid,
			event.Ppid,
			event.User,
			event.UserPid,
			event.UserPpid,
			event.Comm,
			snoop.Filename,
			event.CgroupName,
			event.CgroupId,
			event.TimestampNsExit,
			snoop.ReturnCode,
			event.TimestampNs,
			event.LatencyNs,
			event.EventType,
			event.NodeName,
			float64(event.LatencyNs)/1_000_000.0,
			event.TimestampUnixMs,
			wallTime, // wall_time_dt as DateTime64(3)
			event.ContainerId,
			event.ContainerImage,
			string(labelsJSON),
		)
		if err != nil {
			logger.Error("Failed to append row to batch", zap.Error(err))
			return err
		}
	}

	if err := batch.Send(); err != nil {
		logger.Error("Failed to send ClickHouse batch", zap.Error(err))
		return err
	}

	logger.Debug("Batch insert successful", zap.Int("events", len(events)))
	return nil

}

func (ch *Chconnection) insertNetworkEvent(ctx context.Context,events []*pb.EbpfEvent) error{
  logger := logutil.GetLogger()

	batch, err := ch.conn.PrepareBatch(ctx, `
		INSERT INTO audit.network_events (
			pid, uid, gid, ppid, user_pid, user_ppid, cgroup_id, cgroup_name, comm,
			sa_family, saddr_ipv4, daddr_ipv4, sport, dport,saddr_ipv6,daddr_ipv6,
			monotonic_ts_enter_ns, monotonic_ts_exit_ns, return_code, latency_ns,
			event_type, node_name, user, latency_ms, wall_time_ms, wall_time_dt,
			container_id, container_image, container_labels_json, resolved_domain
		)
	`)
	if err != nil {
		logger.Error("Failed to prepare ClickHouse batch", zap.Error(err))
		return err
	}

  for _, event := range events {
    networkPayload, ok := event.Payload.(*pb.EbpfEvent_Network)
		if !ok {
			logger.Warn("Skipping event: unexpected payload type", zap.String("event_type", event.EventType))
			continue
		}
		network := networkPayload.Network
		labelsJSON, _ := json.Marshal(event.ContainerLabelsJson)
		wallTime := time.Unix(0, int64(event.TimestampUnixMs)*int64(time.Millisecond))

		err := batch.Append(
			event.Pid,
			event.Uid,
			event.Gid,
			event.Ppid,
			event.UserPid,
			event.UserPpid,
			event.CgroupId,
			event.CgroupName,
			event.Comm,
			network.SaFamily,
			network.Saddrv4,
			network.Daddrv4,
			network.Sport,
			network.Dport,
      network.Saddrv6,
      network.Daddrv6,
			event.TimestampNs,
			event.TimestampNsExit,
			network.ReturnCode,
			event.LatencyNs,
			event.EventType,
			event.NodeName,
			event.User,
			float64(event.LatencyNs)/1_000_000.0,
			event.TimestampUnixMs,
			wallTime,
			event.ContainerId,
			event.ContainerImage,
			string(labelsJSON),
      network.ResolvedDomain,
		)
		if err != nil {
			logger.Error("Failed to append network row", zap.Error(err))
			return err
		}
	}

	return batch.Send()
}

func (ch *Chconnection) insertPtraceEvent(ctx context.Context,events []*pb.EbpfEvent) error{
  logger := logutil.GetLogger()

	batch, err := ch.conn.PrepareBatch(ctx, `
		INSERT INTO audit.ptrace_events (
			pid, uid, gid, ppid, user_pid, user_ppid, cgroup_id, cgroup_name, comm,
			request, target_pid, addr, data, request_name,
			monotonic_ts_enter_ns, monotonic_ts_exit_ns, return_code, latency_ns,
			event_type, node_name, user,
			latency_ms, wall_time_ms, wall_time_dt,
			container_id, container_image, container_labels_json
		)
	`)

  if err != nil {
		logger.Error("Failed to prepare ClickHouse ptrace batch", zap.Error(err))
		return err
	}

  for _, event := range events {
		ptracePayload, ok := event.Payload.(*pb.EbpfEvent_Ptrace)
		if !ok {
			logger.Warn("Skipping event: unexpected payload type", zap.String("event_type", event.EventType))
			continue
		}
		ptrace := ptracePayload.Ptrace
		labelsJSON, _ := json.Marshal(event.ContainerLabelsJson)
		wallTime := time.Unix(0, int64(event.TimestampUnixMs)*int64(time.Millisecond))

		err := batch.Append(
			event.Pid,
			event.Uid,
			event.Gid,
			event.Ppid,
			event.UserPid,
			event.UserPpid,
			event.CgroupId,
			event.CgroupName,
			event.Comm,

			ptrace.Request,
			ptrace.TargetPid,
			ptrace.Addr,
			ptrace.Data,
			ptrace.RequestName,

			event.TimestampNs,
			event.TimestampNsExit,
			ptrace.ReturnCode,
			event.LatencyNs,

			event.EventType,
			event.NodeName,
			event.User,

			float64(event.LatencyNs)/1_000_000.0,
			event.TimestampUnixMs,
			wallTime,

			event.ContainerId,
			event.ContainerImage,
			string(labelsJSON),
		)
		if err != nil {
			logger.Error("Failed to append ptrace row", zap.Error(err))
			return err
		}
	}
  
  return batch.Send()
}

func (ch *Chconnection) insertMmapEvent(ctx context.Context,events []*pb.EbpfEvent) error{
  logger := logutil.GetLogger()

	batch, err := ch.conn.PrepareBatch(ctx, `
		INSERT INTO audit.mmap_events (
			pid, uid, gid, ppid, user_pid, user_ppid, cgroup_id, cgroup_name, comm,
			addr, len, prot, flags, fd, off,
			monotonic_ts_enter_ns, monotonic_ts_exit_ns, return_code, latency_ns,
			event_type, node_name, user,
			latency_ms, wall_time_ms, wall_time_dt,
			container_id, container_image, container_labels_json
		)
	`)

  if err != nil {
		logger.Error("Failed to prepare ClickHouse ptrace batch", zap.Error(err))
		return err
	}

  for _, event := range events {
		mmapPayload, ok := event.Payload.(*pb.EbpfEvent_Mmap)
		if !ok {
			logger.Warn("Skipping event: unexpected payload type", zap.String("event_type", event.EventType))
			continue
		}
		mmap := mmapPayload.Mmap
		labelsJSON, _ := json.Marshal(event.ContainerLabelsJson)
		wallTime := time.Unix(0, int64(event.TimestampUnixMs)*int64(time.Millisecond))

		err := batch.Append(
			event.Pid,
			event.Uid,
			event.Gid,
			event.Ppid,
			event.UserPid,
			event.UserPpid,
			event.CgroupId,
			event.CgroupName,
			event.Comm,

			mmap.Addr,
			mmap.Len,
			mmap.Prot,
			mmap.Flags,
			mmap.Fd,
      mmap.Off,

			event.TimestampNs,
			event.TimestampNsExit,
			mmap.ReturnCode,
			event.LatencyNs,

			event.EventType,
			event.NodeName,
			event.User,

			float64(event.LatencyNs)/1_000_000.0,
			event.TimestampUnixMs,
			wallTime,

			event.ContainerId,
			event.ContainerImage,
			string(labelsJSON),
		)
		if err != nil {
			logger.Error("Failed to append ptrace row", zap.Error(err))
			return err
		}
	}
  
  return batch.Send()
}


func (ch *Chconnection) insertMountEvent(ctx context.Context,events []*pb.EbpfEvent) error{
  logger := logutil.GetLogger()

	batch, err := ch.conn.PrepareBatch(ctx, `
		INSERT INTO audit.mount_events (
			pid, uid, gid, ppid, user_pid, user_ppid, cgroup_id, cgroup_name, comm,
			dev_name, dir_name, type, flags,
			monotonic_ts_enter_ns, monotonic_ts_exit_ns, return_code, latency_ns,
			event_type, node_name, user,
			latency_ms, wall_time_ms, wall_time_dt,
			container_id, container_image, container_labels_json
		)
	`)

  if err != nil {
		logger.Error("Failed to prepare ClickHouse ptrace batch", zap.Error(err))
		return err
	}

  for _, event := range events {
		mountPayload, ok := event.Payload.(*pb.EbpfEvent_Mount)
		if !ok {
			logger.Warn("Skipping event: unexpected payload type", zap.String("event_type", event.EventType))
			continue
		}
		mount := mountPayload.Mount
		labelsJSON, _ := json.Marshal(event.ContainerLabelsJson)
		wallTime := time.Unix(0, int64(event.TimestampUnixMs)*int64(time.Millisecond))

		err := batch.Append(
			event.Pid,
			event.Uid,
			event.Gid,
			event.Ppid,
			event.UserPid,
			event.UserPpid,
			event.CgroupId,
			event.CgroupName,
			event.Comm,

			mount.DevName,
			mount.DirName,
			mount.Type,
			mount.Flags,

			event.TimestampNs,
			event.TimestampNsExit,
			mount.ReturnCode,
			event.LatencyNs,

			event.EventType,
			event.NodeName,
			event.User,

			float64(event.LatencyNs)/1_000_000.0,
			event.TimestampUnixMs,
			wallTime,

			event.ContainerId,
			event.ContainerImage,
			string(labelsJSON),
		)
		if err != nil {
			logger.Error("Failed to append ptrace row", zap.Error(err))
			return err
		}
	}
  
  return batch.Send()
}

func (ch *Chconnection) insertResourceEvent(ctx context.Context,events []*pb.EbpfEvent) error{
  logger := logutil.GetLogger()

	batch, err := ch.conn.PrepareBatch(ctx, `
	INSERT INTO audit.resource_events (
		pid, comm, uid, gid, ppid, user_pid, user_ppid, cgroup_id, cgroup_name, user,
		cpu_ns, user_faults, kernel_faults,
		vm_mmap_bytes, vm_munmap_bytes,
		vm_brk_grow_bytes, vm_brk_shrink_bytes,
		bytes_written, bytes_read, isActive,
		wall_time_dt, wall_time_ms,
    container_id, container_image, container_labels_json
	)
  `)

  if err != nil {
    return fmt.Errorf("failed to prepare batch: %w", err)
  }
  for _, event := range events {
		resourcePayload, ok := event.Payload.(*pb.EbpfEvent_Resource)
		if !ok {
			logger.Warn("Skipping event: unexpected payload type", zap.String("event_type", event.EventType))
			continue
		}
		resource := resourcePayload.Resource
    labelsJSON, _ := json.Marshal(event.ContainerLabelsJson)
    wallTime := time.Unix(0, int64(event.TimestampUnixMs)*int64(time.Millisecond))

		err := batch.Append(
			event.Pid,
			event.Comm,
      event.Uid,
      event.Gid,
      event.Ppid,
      event.UserPid,
      event.UserPpid,
      event.CgroupId,
      event.CgroupName,
      event.User,
      resource.CpuNs,
      resource.UserFaults,
      resource.KernelFaults,
      resource.VmMmapBytes,
      resource.VmMunmapBytes,
      resource.VmBrkGrowBytes,
      resource.VmBrkShrinkBytes,
      resource.BytesWritten,
      resource.BytesRead,
      resource.IsActive,
      wallTime,
      event.TimestampUnixMs,

      event.ContainerId,
			event.ContainerImage,
			string(labelsJSON),
		)
		if err != nil {
			logger.Error("Failed to append ptrace row", zap.Error(err))
			return err
		}
	}
  
  return batch.Send()
}

func (ch *Chconnection) insertSyacallFreq(ctx context.Context,events []*pb.EbpfEvent) error{
  logger := logutil.GetLogger()

	batch, err := ch.conn.PrepareBatch(ctx, `
	INSERT INTO audit.syscall_freq_events (
		pid, comm, uid, gid, ppid, user_pid, user_ppid, cgroup_id, cgroup_name, user,
		syscall_vector_json,
		wall_time_dt, wall_time_ms,
    container_id, container_image, container_labels_json
	)
  `)

  if err != nil {
    return fmt.Errorf("failed to prepare batch: %w", err)
  }
  for _, event := range events {
		syscallFreqPayload, ok := event.Payload.(*pb.EbpfEvent_SyscallFreqAgg)
		if !ok {
			logger.Warn("Skipping event: unexpected payload type", zap.String("event_type", event.EventType))
			continue
		}
		sysfreq := syscallFreqPayload.SyscallFreqAgg
    labelsJSON, _ := json.Marshal(event.ContainerLabelsJson)
    wallTime := time.Unix(0, int64(event.TimestampUnixMs)*int64(time.Millisecond))

		err := batch.Append(
			event.Pid,
			event.Comm,
      event.Uid,
      event.Gid,
      event.Ppid,
      event.UserPid,
      event.UserPpid,
      event.CgroupId,
      event.CgroupName,
      event.User,
      sysfreq.VectorJson,
      wallTime,
      event.TimestampUnixMs,

      event.ContainerId,
			event.ContainerImage,
			string(labelsJSON),
		)
		if err != nil {
			logger.Error("Failed to append ptrace row", zap.Error(err))
			return err
		}
	}
  
  return batch.Send()
}



func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''") // escape single quotes
}

func isNetworkevent(eventType string) bool{
  switch eventType {
  case "connect", "accept":
		return true
	default:
		return false
	}
}

func isSnoopevent(eventType string) bool{
  switch eventType {
  case "open", "execve","chmod":
		return true
	default:
		return false
	}
}

