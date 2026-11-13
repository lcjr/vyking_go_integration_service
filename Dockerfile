# Build Stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /vyking-service ./cmd/service/main.go

# Run Stage
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /vyking-service /app/vyking-service

EXPOSE 8080

ENTRYPOINT ["/app/vyking-service"]