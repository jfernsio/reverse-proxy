package main

import (
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Backend struct {
	URL       string
	ActiveCon int
	IsAlive   bool
}

var (
	lastRequest = make(map[string]time.Time)
	mu          sync.Mutex
	backends    []Backend
	lcMu        sync.Mutex
)

func healthCheck() {
	client := &http.Client{Timeout: 5 * time.Second}
	for {
		for i := range backends {
			resp, err := client.Get(backends[i].URL + "/health")
			alive := err == nil && resp != nil && resp.StatusCode == 200
			if resp != nil {
				resp.Body.Close()
			}

			if alive != backends[i].IsAlive {
				backends[i].IsAlive = alive
				if alive {
					log.Printf("✅ %s is BACK online", backends[i].URL)
				} else {
					log.Printf("❌ %s is DOWN!", backends[i].URL)
				}
			}
		}
		time.Sleep(15 * time.Second)
	}
}

func getLeastConnBackend() *Backend {
	lcMu.Lock()
	defer lcMu.Unlock()

	var chosen *Backend
	for i := range backends {
		if !backends[i].IsAlive {
			continue
		}
		if chosen == nil || backends[i].ActiveCon < chosen.ActiveCon ||
			(backends[i].ActiveCon == chosen.ActiveCon && rand.Intn(2) == 0) {
			chosen = &backends[i]
		}
	}
	if chosen == nil {
		return nil
	}
	chosen.ActiveCon++
	return chosen
}

func init() {
	backendStr := os.Getenv("BACKENDS")
	if backendStr == "" {
		backendStr = "http://backend:9000,http://backend:9001"
		log.Println("Backend env not set using localhost as fallback")
	} else {
		log.Printf("Using backends from env: %s", backendStr)
	}
	backendUrls := strings.Split(backendStr, ",")
	backends = make([]Backend, 0, len(backendUrls))

	for _, b := range backendUrls {
		b = strings.TrimSpace(b)
		if b != "" {
			backends = append(backends, Backend{URL: b, IsAlive: true})
		}
	}
	if len(backends) == 0 {
		log.Fatal("No valid backends provided!")
	}
	log.Printf("Configured backends:%d", len(backends))
}
func main() {
	go healthCheck()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Rate limiting (1 req/sec per IP)
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		now := time.Now()
		mu.Lock()
		if t, ok := lastRequest[ip]; ok && now.Sub(t) < time.Second {
			mu.Unlock()
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		lastRequest[ip] = now
		mu.Unlock()

		// Select backend
		backend := getLeastConnBackend()
		if backend == nil {
			http.Error(w, "No backends available", http.StatusServiceUnavailable)
			return
		}

		targetURL, err := url.Parse(backend.URL)
		if err != nil {
			lcMu.Lock()
			backend.ActiveCon--
			lcMu.Unlock()
			http.Error(w, "Invalid backend", http.StatusBadGateway)
			return
		}

		log.Printf("→ %s %s %s → %s", ip, r.Method, r.URL.Path, backend.URL)

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error to %s: %v", backend.URL, err)
		}

		// Decrement connection count after request
		defer func() {
			lcMu.Lock()
			backend.ActiveCon--
			lcMu.Unlock()
			log.Printf("← %s finished (Active: %d)", backend.URL, backend.ActiveCon)
		}()

		proxy.ServeHTTP(w, r)
	})

	log.Println("🚀 GoBalancer started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
