package main

import (
    "context"
    "net"
    "net/http"
    "os"
    "os/signal"
    "time"

    "github.com/gorilla/mux"
    "google.golang.org/grpc"

    pb "github.com/IvanNovakovic/SOA_Proj/protos"
    "follower-service/auth"
    grpchandler "follower-service/grpc"
    "follower-service/handler"
    "follower-service/repository"
    
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/sirupsen/logrus"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
    "go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
    "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

var logger = logrus.New()

func initLogger() {
    logger.SetFormatter(&logrus.JSONFormatter{})
    logger.SetOutput(os.Stdout)
    logger.SetLevel(logrus.InfoLevel)
}

func initTracer() func(context.Context) error {
    endpoint := os.Getenv("OTEL_EXPORTER_JAEGER_ENDPOINT")
    if endpoint == "" {
        endpoint = "http://localhost:14268/api/traces"
    }

    logger.WithFields(logrus.Fields{
        "service":  "follower-service",
        "action":   "tracer_init",
        "endpoint": endpoint,
    }).Info("Initializing Jaeger tracer")

    exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
    if err != nil {
        logger.WithFields(logrus.Fields{
            "service": "follower-service",
            "action":  "tracer_init",
            "error":   err.Error(),
        }).Fatal("Failed to create Jaeger exporter")
    }

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exp),
        trace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String("follower-service"),
            semconv.ServiceVersionKey.String("1.0.0"),
        )),
    )
    otel.SetTracerProvider(tp)

    logger.WithFields(logrus.Fields{
        "service": "follower-service",
        "action":  "tracer_init",
    }).Info("Jaeger tracer initialized successfully")

    return tp.Shutdown
}
func main() {
    initLogger()
    
    logger.WithFields(logrus.Fields{
        "service": "follower-service",
        "action":  "startup",
    }).Info("Starting follower-service")

    _, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Initialize tracer
    cleanup := initTracer()
    defer cleanup(context.Background())

    neoURI := os.Getenv("NEO4J_URI")
    if neoURI == "" {
        neoURI = "bolt://neo4j:7687"
    }
    neoUser := os.Getenv("NEO4J_USER")
    if neoUser == "" {
        neoUser = "neo4j"
    }
    neoPass := os.Getenv("NEO4J_PASSWORD")
    if neoPass == "" {
        neoPass = "testtest123"
    }

    logger.WithFields(logrus.Fields{
        "service": "follower-service",
        "action":  "db_connect",
        "uri":     neoURI,
    }).Info("Connecting to Neo4j")

    // Try connecting to Neo4j with retries (Neo4j may take time to become ready)
    var repo *repository.NeoRepository
    var err error
    maxAttempts := 20
    for i := 0; i < maxAttempts; i++ {
        attemptCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
        repo, err = repository.NewNeoRepository(attemptCtx, neoURI, neoUser, neoPass)
        cancel()
        if err == nil {
            logger.WithFields(logrus.Fields{
                "service": "follower-service",
                "action":  "db_connect",
            }).Info("Successfully connected to Neo4j")
            break
        }
        // exponential backoff with cap
        backoff := time.Duration(2*(i+1)) * time.Second
        if backoff > 20*time.Second {
            backoff = 20 * time.Second
        }
        logger.WithFields(logrus.Fields{
            "service": "follower-service",
            "action":  "db_connect",
            "attempt": i + 1,
            "error":   err.Error(),
            "backoff": backoff.String(),
        }).Warn("Neo4j connection attempt failed, retrying...")
        time.Sleep(backoff)
    }
    if err != nil {
        logger.WithFields(logrus.Fields{
            "service": "follower-service",
            "action":  "db_connect",
            "error":   err.Error(),
        }).Fatal("Failed to connect to Neo4j after all attempts")
    }

    r := mux.NewRouter()
    
    // Add OpenTelemetry middleware
    r.Use(otelmux.Middleware("follower-service"))
    
    // Prometheus metrics endpoint
    r.Handle("/metrics", promhttp.Handler()).Methods("GET")
    
    // health endpoint - use a short timeout so health checks don't hang indefinitely
    r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        if repo == nil {
            http.Error(w, "unhealthy", http.StatusServiceUnavailable)
            return
        }
        _, cancel := context.WithTimeout(r.Context(), 3*time.Second)
        defer cancel()
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }).Methods("GET")
    
    // create an auth-protected subrouter for protected routes
    authSub := r.PathPrefix("").Subrouter()
    authSub.Use(func(next http.Handler) http.Handler { return auth.JWTAuthMiddleware(next) })
    
    handler.RegisterRoutes(r, authSub, repo)

    srv := &http.Server{
        Handler:      r,
        Addr:         ":8082",
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
    }

    // Start HTTP server
    go func() {
        logger.WithFields(logrus.Fields{
            "service": "follower-service",
            "action":  "server_start",
            "address": srv.Addr,
        }).Info("Follower service HTTP server started")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.WithFields(logrus.Fields{
                "service": "follower-service",
                "action":  "server_start",
                "error":   err.Error(),
            }).Fatal("Failed to start HTTP server")
        }
    }()

    // Start gRPC server
    grpcPort := os.Getenv("GRPC_PORT")
    if grpcPort == "" {
        grpcPort = "9092"
    }
    lis, err := net.Listen("tcp", ":"+grpcPort)
    if err != nil {
        logger.WithFields(logrus.Fields{
            "service": "follower-service",
            "action":  "grpc_listen",
            "port":    grpcPort,
            "error":   err.Error(),
        }).Fatalf("Failed to listen on gRPC port")
    }

    // Create gRPC server with OpenTelemetry interceptors
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
        grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
    )
    pb.RegisterFollowerServiceServer(grpcServer, grpchandler.NewFollowerServer(repo))

    go func() {
        logger.WithFields(logrus.Fields{
            "service": "follower-service",
            "action":  "grpc_start",
            "port":    grpcPort,
        }).Info("Follower service gRPC server started")
        if err := grpcServer.Serve(lis); err != nil {
            logger.WithFields(logrus.Fields{
                "service": "follower-service",
                "action":  "grpc_start",
                "error":   err.Error(),
            }).Fatalf("Failed to serve gRPC")
        }
    }()

    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt)
    <-stop
    logger.WithFields(logrus.Fields{
        "service": "follower-service",
        "action":  "shutdown",
    }).Info("Shutting down servers gracefully")
    
    // Shutdown gRPC server gracefully
    logger.WithFields(logrus.Fields{
        "service": "follower-service",
        "action":  "shutdown_grpc",
    }).Info("Stopping gRPC server")
    grpcServer.GracefulStop()
    
    // Shutdown HTTP server
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    srv.Shutdown(shutdownCtx)
    repo.Close()
    logger.WithFields(logrus.Fields{
        "service": "follower-service",
        "action":  "shutdown",
    }).Info("Server shutdown complete")
}
