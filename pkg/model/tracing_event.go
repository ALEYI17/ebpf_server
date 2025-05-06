package model

import (
	"ebpf_server/internal/grpc/pb"
	"time"
)

type EnrichedEvent struct {
	// From gRPC message
	PID             uint32 `json:"pid"`
	UID             uint32 `json:"uid"`
	GID             uint32 `json:"gid"`
	User            string `json:"user"`
	PPID            uint32 `json:"ppid"`
	UserPID         uint32 `json:"user_pid"`
	UserPPID        uint32 `json:"user_ppid"`
	CgroupName      string `json:"cgroup_name"`
	CgroupID        uint64 `json:"cgroup_id"`
	Comm            string `json:"comm"`
	Filename        string `json:"filename"`
	TimestampNs     uint64 `json:"timestamp_ns"`
	ExitTimestampNs uint64 `json:"timestamp_ns_exit"`
	ReturnCode      int64  `json:"return_code"`
	LatencyNs       uint64 `json:"latency_ns"`
	EventType       string `json:"event_type"`
	NodeName        string `json:"node_name"`

	// Enriched fields
	TimestampISO     string  `json:"timestamp_iso"`
	ExitTimestampISO string  `json:"exit_timestamp_iso"`
	LatencyMs        float64 `json:"latency_ms"`
}

func EnrichEvent(event *pb.EbpfEvent) *EnrichedEvent {
	return &EnrichedEvent{
		PID:             event.Pid,
		UID:             event.Uid,
		GID:             event.Gid,
		User:            event.User,
		PPID:            event.Ppid,
		UserPID:         event.UserPid,
		UserPPID:        event.UserPpid,
		CgroupName:      event.CgroupName,
		CgroupID:        event.CgroupId,
		Comm:            event.Comm,
		Filename:        event.Filename,
		TimestampNs:     event.TimestampNs,
		ExitTimestampNs: event.TimestampNsExit,
		ReturnCode:      event.ReturnCode,
		LatencyNs:       event.LatencyNs,
		EventType:       event.EventType,
		NodeName:        event.NodeName,

		TimestampISO:     time.Unix(0, int64(event.TimestampNs)).Format(time.RFC3339Nano),
		ExitTimestampISO: time.Unix(0, int64(event.TimestampNsExit)).Format(time.RFC3339Nano),
		LatencyMs:        float64(event.LatencyNs) / 1_000_000,
	}
}
