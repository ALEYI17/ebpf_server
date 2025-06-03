# ebpf_server

## 🐳 Build Docker Image

```bash
make docker-build IMG=<image name>:<image tag>
```
### Build and Push Docker Image

```bash
make docker-build docker-push IMG=<image name>:<image tag>
```

## Run with Docker Compose

```bash
docker compose up -d
```

### Connect Another Container to Compose Network

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
