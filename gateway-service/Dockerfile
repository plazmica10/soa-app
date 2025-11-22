FROM golang:1.20-alpine AS build
RUN apk add --no-cache git
WORKDIR /app
COPY . .
RUN go build -o /gateway ./

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /gateway /gateway
EXPOSE 8080
ENTRYPOINT ["/gateway"]
