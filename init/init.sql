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
  timestamp_ns UInt64,
  timestamp_ns_exit UInt64,
  return_code Int64,
  latency_ns UInt64,
  event_type String,
  node_name String,
  user String,
  latency_ms Float64 
)ENGINE = MergeTree()
ORDER BY timestamp_ns;
