FROM golang:1.24 as builder

WORKDIR /workspace

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ebpf_server cmd/main.go

FROM golang:bookworm

COPY --from=builder /workspace/ebpf_server .

RUN ls

ENTRYPOINT ["./ebpf_server"]
