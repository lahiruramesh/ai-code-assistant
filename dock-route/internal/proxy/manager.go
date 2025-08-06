package proxy

import (
    "fmt"
    "log"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"
    "sync"
)

type Manager struct {
    mu      sync.RWMutex
    proxies map[string]*httputil.ReverseProxy
}

func NewManager() *Manager {
    return &Manager{
        proxies: make(map[string]*httputil.ReverseProxy),
    }
}

func (pm *Manager) AddProxy(subdomain string, targetURL string) error {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    target, err := url.Parse(targetURL)
    if err != nil {
        return fmt.Errorf("invalid target URL: %w", err)
    }
    
    proxy := httputil.NewSingleHostReverseProxy(target)
    
    // Custom director
    originalDirector := proxy.Director
    proxy.Director = func(req *http.Request) {
        originalDirector(req)
        req.Host = target.Host
        req.URL.Host = target.Host
        req.URL.Scheme = target.Scheme
    }
    
    pm.proxies[subdomain] = proxy
    log.Printf("Added proxy for subdomain: %s -> %s", subdomain, targetURL)
    
    return nil
}

func (pm *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    host := r.Host
    parts := strings.Split(host, ".")
    
    var subdomain string
    if len(parts) > 2 {
        subdomain = parts[0]
    } else {
        subdomain = "default"
    }
    
    pm.mu.RLock()
    proxy, found := pm.proxies[subdomain]
    pm.mu.RUnlock()
    
    if !found {
        log.Printf("No proxy found for subdomain: %s (Host: %s)", subdomain, host)
        http.Error(w, "Not Found: No application configured for this subdomain.", http.StatusNotFound)
        return
    }
    
    log.Printf("Proxying request for %s to target for subdomain %s", r.URL.String(), subdomain)
    proxy.ServeHTTP(w, r)
}

func (pm *Manager) RemoveProxy(subdomain string) {
    pm.mu.Lock()
    defer pm.mu.Unlock()
    
    delete(pm.proxies, subdomain)
    log.Printf("Removed proxy for subdomain: %s", subdomain)
}

func (pm *Manager) GetActiveSubdomains() []string {
    pm.mu.RLock()
    defer pm.mu.RUnlock()
    
    var subdomains []string
    for subdomain := range pm.proxies {
        subdomains = append(subdomains, subdomain)
    }
    
    return subdomains
}

func (pm *Manager) HasProxy(subdomain string) bool {
    pm.mu.RLock()
    defer pm.mu.RUnlock()
    
    _, exists := pm.proxies[subdomain]
    return exists
}
