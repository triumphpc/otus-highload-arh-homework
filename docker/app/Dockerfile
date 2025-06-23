FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4
RUN swag init -g ./cmd/app/main.go -o ./docs

RUN CGO_ENABLED=0 GOOS=linux go build -o /social-app ./cmd/app/main.go

FROM alpine:latest

COPY --from=builder /social-app /social-app
COPY --from=builder /app/docs /docs
COPY --from=builder /app/migrations /migrations
EXPOSE 8080

CMD ["/social-app"]