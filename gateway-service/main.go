package main

import (
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"
    "time"
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

func main() {
    // route prefix -> backend service (docker service name + internal port)
    routes := map[string]string{
        // blog service: blogs, comments, likes
        "/blog":      "http://blog-service:8081",
        "/blogs":     "http://blog-service:8081",
        "/comments":  "http://blog-service:8081",
        "/likes":     "http://blog-service:8081",

        // follower service
        "/follow":    "http://follower-service:8082",
        "/followers": "http://follower-service:8082",

        // tour service
        "/tour":      "http://tour-service:8083",
        "/tours":     "http://tour-service:8083",
        "/review":    "http://tour-service:8083",
        "/reviews":   "http://tour-service:8083",
        "/kp":        "http://tour-service:8083",

        // stakeholders (users/auth)
		"/health":   "http://stakeholders-service:8080",
        "/user":      "http://stakeholders-service:8080",
        "/users":     "http://stakeholders-service:8080",
        "/auth":      "http://stakeholders-service:8080",
        "/login":     "http://stakeholders-service:8080",
        "/register":  "http://stakeholders-service:8080",
    }

    // build proxies per target to reuse
    proxies := map[string]*httputil.ReverseProxy{}
    for _, target := range routes {
        if _, ok := proxies[target]; !ok {
            proxies[target] = newProxy(target)
        }
    }

    handler := func(w http.ResponseWriter, r *http.Request) {
        path := r.URL.Path
        // find longest matching prefix
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
