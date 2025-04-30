CREATE DATABASE IF NOT EXISTS audit;

CREATE TABLE IF NOT EXISTS audit.tracing_events (
  pid UInt32,
  uid UInt32,
  comm String,
  filename String,
  timestamp_ns_exit UInt64,
  return_code Int64,
  timestamp_ns UInt64,
  latency_ns UInt64,
  node_name String
) ENGINE = MergeTree()
ORDER BY timestamp_ns;
