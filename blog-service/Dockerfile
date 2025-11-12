# Build a minimal Docker image for the blog-service
FROM golang:1.20-alpine AS build
RUN apk add --no-cache git ca-certificates tzdata
WORKDIR /src

# Copy go mod & sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/blog-service ./

FROM alpine:3.18
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /bin/blog-service /bin/blog-service
EXPOSE 8081
ENTRYPOINT ["/bin/blog-service"]
