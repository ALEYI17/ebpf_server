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
