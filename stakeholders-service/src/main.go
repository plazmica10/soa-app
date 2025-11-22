package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"

	pb "github.com/IvanNovakovic/SOA_Proj/protos"
	grpchandler "stakeholders-service/grpc"
	"stakeholders-service/handler"
	"stakeholders-service/repository"

	"github.com/gorilla/mux"
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
		"service":  "stakeholders-service",
		"action":   "tracer_init",
		"endpoint": endpoint,
	}).Info("Initializing Jaeger tracer")

	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service": "stakeholders-service",
			"action":  "tracer_init",
			"error":   err.Error(),
		}).Fatal("Failed to create Jaeger exporter")
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exp),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("stakeholders-service"),
			semconv.ServiceVersionKey.String("1.0.0"),
		)),
	)
	otel.SetTracerProvider(tp)

	logger.WithFields(logrus.Fields{
		"service": "stakeholders-service",
		"action":  "tracer_init",
	}).Info("Jaeger tracer initialized successfully")

	return tp.Shutdown
}

func initDB(ctx context.Context) *mongo.Client {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	clientOpts := options.Client().ApplyURI(uri)

	logger.WithFields(logrus.Fields{
		"service": "stakeholders-service",
		"action":  "db_connect",
		"uri":     uri,
	}).Info("Connecting to MongoDB")

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service": "stakeholders-service",
			"action":  "db_connect",
			"error":   err.Error(),
		}).Fatal("Failed to connect to MongoDB")
	}
	if err := client.Ping(ctx, nil); err != nil {
		logger.WithFields(logrus.Fields{
			"service": "stakeholders-service",
			"action":  "db_ping",
			"error":   err.Error(),
		}).Fatal("Failed to ping MongoDB")
	}

	logger.WithFields(logrus.Fields{
		"service": "stakeholders-service",
		"action":  "db_connect",
	}).Info("Successfully connected to MongoDB")

	return client
}

func main() {
	initLogger()
	
	logger.WithFields(logrus.Fields{
		"service": "stakeholders-service",
		"action":  "startup",
	}).Info("Starting stakeholders-service")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize tracer
	cleanup := initTracer()
	defer cleanup(context.Background())

	client := initDB(ctx)
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "stakeholders"
	}

	repo := repository.NewUserRepository(client.Database(dbName))
	router := mux.NewRouter()
	
	// Add OpenTelemetry middleware
	router.Use(otelmux.Middleware("stakeholders-service"))
	
	// Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// public
	handler.RegisterAuthRoutes(router, repo)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// protected user routes
	protected := router.PathPrefix("").Subrouter()
	handler.RegisterRoutes(protected, repo)
	protected.Use(handler.JWTAuthMiddleware)

	srv := &http.Server{
		Handler: router,
		Addr: ":" + func() string {
			if p := os.Getenv("PORT"); p != "" {
				return p
			}
			return "8080"
		}(),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// Start HTTP server
	go func() {
		logger.WithFields(logrus.Fields{
			"service": "stakeholders-service",
			"action":  "server_start",
			"address": srv.Addr,
		}).Info("Stakeholders service HTTP server started")
		if err := srv.ListenAndServe(); err != nil {
			logger.WithFields(logrus.Fields{
				"service": "stakeholders-service",
				"action":  "server_start",
				"error":   err.Error(),
			}).Fatal("Failed to start HTTP server")
		}
	}()

	// Start gRPC server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service": "stakeholders-service",
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
	pb.RegisterStakeholderServiceServer(grpcServer, grpchandler.NewStakeholderServer(repo))

	go func() {
		logger.WithFields(logrus.Fields{
			"service": "stakeholders-service",
			"action":  "grpc_start",
			"port":    grpcPort,
		}).Info("Stakeholders service gRPC server started")
		if err := grpcServer.Serve(lis); err != nil {
			logger.WithFields(logrus.Fields{
				"service": "stakeholders-service",
				"action":  "grpc_start",
				"error":   err.Error(),
			}).Fatalf("Failed to serve gRPC")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	logger.WithFields(logrus.Fields{
		"service": "stakeholders-service",
		"action":  "shutdown",
	}).Info("Shutting down servers gracefully")
	
	// Shutdown gRPC server gracefully
	logger.WithFields(logrus.Fields{
		"service": "stakeholders-service",
		"action":  "shutdown_grpc",
	}).Info("Stopping gRPC server")
	grpcServer.GracefulStop()
	
	// Shutdown HTTP server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
	client.Disconnect(shutdownCtx)
	logger.WithFields(logrus.Fields{
		"service": "stakeholders-service",
		"action":  "shutdown",
	}).Info("Server shutdown complete")
}
