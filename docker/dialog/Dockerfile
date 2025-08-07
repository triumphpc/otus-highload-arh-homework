FROM golang:1.24-alpine AS builder

# Установка grpc_health_probe
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.24 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o dialog ./cmd/dialog

FROM alpine:latest

# Установка необходимых пакетов
RUN apk --no-cache add ca-certificates && \
    GRPC_HEALTH_PROBE_VERSION=v0.4.24 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

COPY --from=builder /app/dialog .
COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe

CMD ["./dialog"]