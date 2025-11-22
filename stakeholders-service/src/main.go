package main

import (
	"context"
	"log"
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
)

func initDB(ctx context.Context) *mongo.Client {
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}
	clientOpts := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal("mongo connect:", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("mongo ping:", err)
	}

	return client
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := initDB(ctx)
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "stakeholders"
	}

	repo := repository.NewUserRepository(client.Database(dbName))
	router := mux.NewRouter()

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
		log.Println("HTTP server started on " + srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// Start gRPC server
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen on gRPC port: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterStakeholderServiceServer(grpcServer, grpchandler.NewStakeholderServer(repo))

	go func() {
		log.Printf("gRPC server started on :%s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("shutting down servers...")
	
	// Shutdown gRPC server gracefully
	grpcServer.GracefulStop()
	
	// Shutdown HTTP server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
	client.Disconnect(shutdownCtx)
}
