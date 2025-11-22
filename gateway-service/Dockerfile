FROM golang:1.22-alpine AS build
RUN apk add --no-cache git
WORKDIR /app
# Copy protos directory for local module replacement
COPY protos /protos
# Copy gateway-service files
COPY gateway-service/ ./
RUN go build -o /gateway ./

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /gateway /gateway
EXPOSE 8080
ENTRYPOINT ["/gateway"]
