package programs

import (
	"context"
	"ebpf_server/internal/grpc/pb"
)

type Load_tracer interface {
	Close()
	Run(context.Context, string) <-chan *pb.EbpfEvent
}

type Load_tracer_batch interface {
	Close()
	Run(context.Context, string) <-chan *pb.Batch
}


const (
	LoaderOpen     = "open"
	Loaderexecve = "execve"
  LoaderChmod = "chmod"
  LoaderConnect = "connect"
  LoaderAccept = "accept"
  LoaderPtrace = "ptrace"
  LoaderMmap = "mmap"
  LoaderMount = "mount"
  LoadUmount = "umount"
  LoadResource = "resource"
  LoadSyscallFreq = "syscallFreq"
)

