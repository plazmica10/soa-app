package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"
    "time"

    pb "github.com/IvanNovakovic/SOA_Proj/protos"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func newProxy(target string) *httputil.ReverseProxy {
    u, err := url.Parse(target)
    if err != nil {
        log.Fatalf("invalid proxy target %s: %v", target, err)
    }
    proxy := httputil.NewSingleHostReverseProxy(u)
    // tweak the director to preserve original host header if needed
    originalDirector := proxy.Director
    proxy.Director = func(req *http.Request) {
        originalDirector(req)
        req.Host = u.Host
    }
    proxy.ModifyResponse = func(resp *http.Response) error {
        // add small header to indicate proxied
        resp.Header.Set("X-Gateway-Proxy", "gateway-service")
        return nil
    }
    proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
        log.Printf("proxy error for %s -> %v", req.URL.Path, err)
        http.Error(rw, "Bad Gateway", http.StatusBadGateway)
    }
    return proxy
}

// gRPC client connections
type grpcClients struct {
    stakeholderConn *grpc.ClientConn
    stakeholderClient pb.StakeholderServiceClient
    followerConn *grpc.ClientConn
    followerClient pb.FollowerServiceClient
}

func initGRPCClients() (*grpcClients, error) {
    // Connect to stakeholder service
    stakeholderConn, err := grpc.Dial("stakeholders-service:9090", 
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, err
    }

    // Connect to follower service
    followerConn, err := grpc.Dial("follower-service:9092", 
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        stakeholderConn.Close()
        return nil, err
    }

    return &grpcClients{
        stakeholderConn: stakeholderConn,
        stakeholderClient: pb.NewStakeholderServiceClient(stakeholderConn),
        followerConn: followerConn,
        followerClient: pb.NewFollowerServiceClient(followerConn),
    }, nil
}

func (gc *grpcClients) Close() {
    if gc.stakeholderConn != nil {
        gc.stakeholderConn.Close()
    }
    if gc.followerConn != nil {
        gc.followerConn.Close()
    }
}

func main() {
    // Initialize gRPC clients
    grpcClients, err := initGRPCClients()
    if err != nil {
        log.Fatalf("failed to init gRPC clients: %v", err)
    }
    defer grpcClients.Close()

    // route prefix -> backend service (docker service name + internal port)
    routes := map[string]string{
        // blog service: blogs, comments, likes
        "/blog":      "http://blog-service:8081",
        "/blogs":     "http://blog-service:8081",
        "/comments":  "http://blog-service:8081",
        "/likes":     "http://blog-service:8081",

        // follower service (other endpoints still via HTTP)
        "/follow":    "http://follower-service:8082",
        "/following": "http://follower-service:8082",
        "/recommendations": "http://follower-service:8082",

        // tour service
        "/tour":      "http://tour-service:8083",
        "/tours":     "http://tour-service:8083",
        "/review":    "http://tour-service:8083",
        "/reviews":   "http://tour-service:8083",
        "/kp":        "http://tour-service:8083",

        // stakeholders (users/auth) - register still via HTTP
		"/health":   "http://stakeholders-service:8080",
        "/user":      "http://stakeholders-service:8080",
        "/users":     "http://stakeholders-service:8080",
        "/auth":      "http://stakeholders-service:8080",
        "/register":  "http://stakeholders-service:8080",
    }

    // build proxies per target to reuse
    proxies := map[string]*httputil.ReverseProxy{}
    for _, target := range routes {
        if _, ok := proxies[target]; !ok {
            proxies[target] = newProxy(target)
        }
    }

    // gRPC handler for login - define before use
    handleLogin := func(w http.ResponseWriter, r *http.Request, client pb.StakeholderServiceClient) {
        if r.Method != "POST" {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var req struct {
            Username string `json:"username"`
            Password string `json:"password"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "invalid json", http.StatusBadRequest)
            return
        }

        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()

        resp, err := client.Login(ctx, &pb.LoginRequest{
            Username: req.Username,
            Password: req.Password,
        })
        if err != nil {
            log.Printf("gRPC login error: %v", err)
            http.Error(w, "login failed", http.StatusUnauthorized)
            return
        }

        // Return response in format compatible with existing clients
        result := map[string]interface{}{
            "access_token": resp.Token,
            "token_type":   "Bearer",
            "expires_in":   3600,
            "user": map[string]interface{}{
                "id": resp.UserId,
            },
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(result)
    }

    // gRPC handler for get followers
    handleGetFollowers := func(w http.ResponseWriter, r *http.Request, client pb.FollowerServiceClient) {
        if r.Method != "GET" {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        // Extract user ID from path /followers/{id}
        parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
        if len(parts) < 2 {
            http.Error(w, "user id required", http.StatusBadRequest)
            return
        }
        userID := parts[1]

        ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
        defer cancel()

        resp, err := client.GetFollowers(ctx, &pb.GetFollowersRequest{
            UserId: userID,
        })
        if err != nil {
            log.Printf("gRPC get followers error: %v", err)
            http.Error(w, "failed to get followers", http.StatusInternalServerError)
            return
        }

        // Convert to string array format
        var followers []string
        for _, f := range resp.Followers {
            followers = append(followers, f.UserId)
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(followers)
    }

    // Main request handler
    handler := func(w http.ResponseWriter, r *http.Request) {
        path := r.URL.Path

        // Handle gRPC-based endpoints
        if path == "/login" || path == "/auth/login" {
            handleLogin(w, r, grpcClients.stakeholderClient)
            return
        }

        if strings.HasPrefix(path, "/followers/") {
            handleGetFollowers(w, r, grpcClients.followerClient)
            return
        }

        // find longest matching prefix for HTTP proxy
        var matchedTarget string
        var matchedPrefix string
        for prefix, target := range routes {
            if strings.HasPrefix(path, prefix) {
                if len(prefix) > len(matchedPrefix) {
                    matchedPrefix = prefix
                    matchedTarget = target
                }
            }
        }
        if matchedTarget != "" {
            proxies[matchedTarget].ServeHTTP(w, r)
            return
        }

        // no prefix match -> 502 or simple routing by host/path
        http.Error(w, "Service not found", http.StatusBadGateway)
    }

    srv := &http.Server{
        Addr:         ":8080",
        Handler:      http.HandlerFunc(handler),
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 30 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    log.Printf("gateway service starting on %s", srv.Addr)
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("gateway listen error: %v", err)
    }
}
