FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o feed-updater ./cmd/feed-updater

FROM alpine:latest
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/feed-updater .

CMD ["./feed-updater"]