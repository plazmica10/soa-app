# Multi-stage Dockerfile for tour-service
FROM golang:1.22-alpine AS builder
WORKDIR /src

# Copy module files first to leverage cache
COPY tour-service/go.mod tour-service/go.sum ./
RUN go mod download

COPY tour-service/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags='-s -w' -o /out/tour-service ./

FROM alpine:3.18 AS runtime
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/tour-service /usr/local/bin/tour-service

EXPOSE 8083
ENTRYPOINT ["/usr/local/bin/tour-service"]
