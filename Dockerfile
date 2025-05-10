FROM golang:1.24 as builder

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ebpf_server cmd/main.go

ENTRYPOINT ["./ebpf_server"]
