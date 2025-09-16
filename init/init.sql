CREATE DATABASE IF NOT EXISTS audit;

CREATE TABLE IF NOT EXISTS audit.tracing_events (
  pid UInt32,
  uid UInt32,
  gid UInt32,
  ppid UInt32,
  user_pid UInt32,
  user_ppid UInt32,
  cgroup_id UInt64,
  cgroup_name String,
  comm String,
  filename String,
  monotonic_ts_enter_ns UInt64,
  monotonic_ts_exit_ns UInt64,
  return_code Int64,
  latency_ns UInt64,
  event_type String,
  node_name String,
  user String,
  latency_ms Float64, 
  wall_time_ms Int64,
  wall_time_dt DateTime64(3),
  container_id String,
  container_image String,
  container_labels_json JSON

)ENGINE = MergeTree()
ORDER BY wall_time_ms;

CREATE TABLE IF NOT EXISTS audit.network_events (
  pid UInt32,
  uid UInt32,
  gid UInt32,
  ppid UInt32,
  user_pid UInt32,
  user_ppid UInt32,
  cgroup_id UInt64,
  cgroup_name String,
  comm String,
  
  sa_family String,               -- socket family (e.g. AF_INET, AF_INET6, etc.)
  saddr_ipv4 String,             -- source IPv4 address as string
  daddr_ipv4 String,             -- dest IPv4 address as string
  sport String,                 -- source port
  dport String,                 -- dest port
  saddr_ipv6 String,  
  daddr_ipv6 String,
  resolved_domain Nullable(String),
  monotonic_ts_enter_ns UInt64,
  monotonic_ts_exit_ns UInt64,
  return_code Int64,
  latency_ns UInt64,

  event_type String,
  node_name String,
  user String,

  latency_ms Float64, 
  wall_time_ms Int64,
  wall_time_dt DateTime64(3),

  container_id String,
  container_image String,
  container_labels_json JSON
) ENGINE = MergeTree()
ORDER BY wall_time_ms;

CREATE TABLE IF NOT EXISTS audit.ptrace_events (
  pid UInt32,
  uid UInt32,
  gid UInt32,
  ppid UInt32,
  user_pid UInt32,
  user_ppid UInt32,
  cgroup_id UInt64,
  cgroup_name String,
  comm String,
  
  request Int64,
  target_pid Int64,
  addr UInt64,
  data UInt64,
  request_name String,
  monotonic_ts_enter_ns UInt64,
  monotonic_ts_exit_ns UInt64,
  return_code Int64,
  latency_ns UInt64,

  event_type String,
  node_name String,
  user String,

  latency_ms Float64, 
  wall_time_ms Int64,
  wall_time_dt DateTime64(3),

  container_id String,
  container_image String,
  container_labels_json JSON
) ENGINE = MergeTree()
ORDER BY wall_time_ms;

CREATE TABLE IF NOT EXISTS audit.mmap_events (
  pid UInt32,
  uid UInt32,
  gid UInt32,
  ppid UInt32,
  user_pid UInt32,
  user_ppid UInt32,
  cgroup_id UInt64,
  cgroup_name String,
  comm String,
  
  addr UInt64,
  len UInt64,
  prot UInt64,
  flags UInt64,
  fd UInt64,
  off UInt64,

  monotonic_ts_enter_ns UInt64,
  monotonic_ts_exit_ns UInt64,
  return_code Int64,
  latency_ns UInt64,

  event_type String,
  node_name String,
  user String,

  latency_ms Float64, 
  wall_time_ms Int64,
  wall_time_dt DateTime64(3),

  container_id String,
  container_image String,
  container_labels_json JSON
) ENGINE = MergeTree()
ORDER BY wall_time_ms;

CREATE TABLE IF NOT EXISTS audit.mount_events (
  pid UInt32,
  uid UInt32,
  gid UInt32,
  ppid UInt32,
  user_pid UInt32,
  user_ppid UInt32,
  cgroup_id UInt64,
  cgroup_name String,
  comm String,
  
  dev_name String,
  dir_name String,
  type String,
  flags UInt64,

  monotonic_ts_enter_ns UInt64,
  monotonic_ts_exit_ns UInt64,
  return_code Int64,
  latency_ns UInt64,

  event_type String,
  node_name String,
  user String,

  latency_ms Float64, 
  wall_time_ms Int64,
  wall_time_dt DateTime64(3),

  container_id String,
  container_image String,
  container_labels_json JSON
) ENGINE = MergeTree()
ORDER BY wall_time_ms;

CREATE TABLE IF NOT EXISTS audit.resource_events (
  pid UInt32,
  comm String,

  uid UInt32,
  gid UInt32,
  ppid UInt32,
  user_pid UInt32,
  user_ppid UInt32,
  cgroup_id UInt64,
  cgroup_name String,
  user String,

  cpu_ns Float64,
  user_faults UInt64,
  kernel_faults UInt64,
  vm_mmap_bytes Float64,
  vm_munmap_bytes Float64,
  vm_brk_grow_bytes Float64,
  vm_brk_shrink_bytes Float64,
  bytes_written Float64,
  bytes_read Float64,
  isActive UInt32,

  wall_time_dt DateTime64(3),
  wall_time_ms Int64,
  
  container_id String,
  container_image String,
  container_labels_json JSON
  
) ENGINE = MergeTree()
ORDER BY wall_time_ms;

CREATE TABLE IF NOT EXISTS audit.syscall_freq_events (
  -- Common process / container metadata
  pid UInt32,
  comm String,

  uid UInt32,
  gid UInt32,
  ppid UInt32,
  user_pid UInt32,
  user_ppid UInt32,
  cgroup_id UInt64,
  cgroup_name String,
  user String,

  -- Syscall aggregation
  syscall_vector_json JSON,    -- e.g. {"0":12,"1":4,"60":9}
  distinct_syscalls UInt32,    -- number of syscalls seen in this batch
  total_syscalls UInt64,       -- sum of all counts in the vector

  -- Timestamps
  wall_time_dt DateTime64(3),
  wall_time_ms Int64,
  
  -- Container metadata
  container_id String,
  container_image String,
  container_labels_json JSON
)
ENGINE = MergeTree()
ORDER BY wall_time_ms;
