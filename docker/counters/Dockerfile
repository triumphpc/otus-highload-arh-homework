FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4
RUN swag init -g ./cmd/counters/main.go -o ./docs/counters

RUN CGO_ENABLED=0 GOOS=linux go build -o /social-app ./cmd/counters/main.go

FROM alpine:latest
COPY --from=builder /social-app /social-app
COPY --from=builder /app/docs /docs/counters
EXPOSE 8081

CMD ["/social-app"]