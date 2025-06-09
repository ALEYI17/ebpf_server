# InfraSight Server (ebpf_server)

The **eBPF Server** is a core component of the [InfraSight](https://github.com/ALEYI17/InfraSight) observability platform. It receives telemetry from `ebpf_loader` agents running on nodes, optionally enriches the data, and stores it efficiently in **ClickHouse**. It is designed to be fast, modular, and ready for production use.

## 📦 Features

- gRPC server for ingesting eBPF telemetry from multiple nodes
- Event enrichment (e.g., timestamp formatting, latency normalization)
- ClickHouse ingestion using batching

## 🧱 Technologies Used and Dependencies
- [Go](https://golang.org/) (>= 1.21)
- [gRPC](https://grpc.io/)
- [Protocol Buffers](https://protobuf.dev/)
- [ClickHouse](https://clickhouse.com)
- [clickhouse-go v2](https://github.com/ClickHouse/clickhouse-go)

## 🚀 Getting Started
### 🐳 Build Docker Image

```bash
make docker-build IMG=<image name>:<image tag>
```
### Build and Push Docker Image

```bash
make docker-build docker-push IMG=<image name>:<image tag>
```

### 🐳 Docker Compose
```bash
docker compose up -d
```

#### Connect Another Container to Compose Network

```bash
docker network ls
```
then:

```bash
docker run -it \
  -e DB_ADDRESS=clickhouse:9000 \
  --network <network> \
  <image name>:<image tag>
```
Replace `<network>` , `<image name>`, and `<image tag>` accordingly.

## 🛠️ Building from Source

### Clone the repository
```bash
git clone https://github.com/ALEYI17/ebpf_server.git
cd ebpf_server
```

### Compile the Go code
```bash
go build -o ebpf-server ./cmd/main.go
```
### 🧪 Compiling Protobuf
If you modify the `.proto` file, recompile Go stubs:

```bash
cd internal/grpc/pb && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ebpf_event.proto
```

## ⚙️ Configuration

You can configure the server using environment variables:

| Variable         | Description                              | Default          |
| ---------------- | ---------------------------------------- | ---------------- |
| `SERVER_PORT`    | gRPC port to receive events              | `8080`           |
| `DB_ADDRESS`     | ClickHouse address                       | `localhost:9000` |
| `DB_USER`        | ClickHouse user                          | `user`           |
| `DB_PASSWORD`    | ClickHouse password                      | `password`       |
| `DB_NAME`        | ClickHouse database name                 | `audit`          |
| `BATCH_MAX_SIZE` | Max number of events per insert batch    | `1000`           |
| `BATCH_FLUSH_MS` | Max time (in ms) before flushing a batch | `10000`          |

Example:

```bash
export SERVER_PORT=8080
export DB_ADDRESS=localhost:9000
export DB_USER=default
export DB_PASSWORD=yourpassword
export DB_NAME=audit
export BATCH_MAX_SIZE=1000
export BATCH_FLUSH_MS=5000

go run cmd/main.go
```

## 🗃️ ClickHouse Schema

Ensure your database has the following table:

```sql
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
  
  sa_family String,               
  saddr_ipv4 String,             
  daddr_ipv4 String,             
  sport String,                 
  dport String,                 
  saddr_ipv6 String,  
  daddr_ipv6 String,
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
```
> ⚠️ Note: Can find the sql in the `./init/init.sql`

## 🔐 Security Note

The current server does **not** enforce mTLS or authentication. For production deployments, it is recommended to:

* Enable mTLS for gRPC communication
* Lock down ClickHouse access
* Enable rate limiting and retries

---

## 📚 Related Repositories

This is part of the **[InfraSight](https://github.com/ALEYI17/InfraSight)** platform:

- [`infrasight-controller`](https://github.com/ALEYI17/infrasight-controller): Kubernetes controller to manage agents
- [`ebpf_loader`](https://github.com/ALEYI17/ebpf_loader): Agent that collects and sends eBPF telemetry from nodes
- [`ebpf_server`](https://github.com/ALEYI17/ebpf_server): Receives and stores events (e.g., to ClickHouse)
- [`ebpf_deploy`](https://github.com/ALEYI17/ebpf_deploy): Helm charts to deploy the stack
