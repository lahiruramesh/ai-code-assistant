package proxy

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "time"
)

type Server struct {
    manager *Manager
    server  *http.Server
    port    string
}

func NewServer(port string) *Server {
    manager := NewManager()
    
    server := &http.Server{
        Addr:              ":" + port,
        Handler:           manager,
        ReadHeaderTimeout: 5 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       15 * time.Second,
    }
    
    return &Server{
        manager: manager,
        server:  server,
        port:    port,
    }
}

func (s *Server) AddProxy(subdomain, targetURL string) error {
    return s.manager.AddProxy(subdomain, targetURL)
}

func (s *Server) RemoveProxy(subdomain string) {
    s.manager.RemoveProxy(subdomain)
}

func (s *Server) Start() error {
    log.Printf("Starting reverse proxy server on port %s", s.port)
    
    if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        return fmt.Errorf("proxy server failed: %w", err)
    }
    
    return nil
}

func (s *Server) Stop(ctx context.Context) error {
    log.Println("Shutting down proxy server...")
    return s.server.Shutdown(ctx)
}

func (s *Server) GetActiveProxies() []string {
    return s.manager.GetActiveSubdomains()
}
