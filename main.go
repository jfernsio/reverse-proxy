package main

import (
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	backends    = []Backend{
		{URL: "http://localhost:9000", IsAlive: true},
		{URL: "http://localhost:9001", IsAlive: true},
	}
	connectionsCount = 0
)

// health check for backends
func healthCheck() {
	for {
		for i := range backends {
			resp, err := http.Get(backends[i].URL)

			if err != nil || resp.StatusCode != 200 {
				backends[i].IsAlive = false
				log.Println(backends[i].URL, "is DOWN!")
			} else {
				backends[i].IsAlive = true
			}
		}
		time.Sleep(30 * time.Second)
	}
}

// simple least connection load balcncer
var lcMu sync.Mutex

func getLeastConnBackend() *Backend {

	lcMu.Lock()
	defer lcMu.Unlock()
	var chosen *Backend
	for i := range backends {
		//choose backend with least active connections, if tie, randomly choose one
		//skip ded backends
		if !backends[i].IsAlive {
			continue
		}
		if chosen == nil || backends[i].ActiveCon < chosen.ActiveCon ||
			(backends[i].ActiveCon == chosen.ActiveCon && rand.Intn(2) == 0) {
			chosen = &backends[i]
		}
	}
	if len(backends) == 0 {
		return nil
	}

	chosen.ActiveCon++
	return chosen

}

func main() {
	go healthCheck()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Extract IP without port
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)

		now := time.Now()

		//apply rate limitin for each IP to 1 req per sec
		mu.Lock()
		if t, ok := lastRequest[ip]; ok {
			if now.Sub(t) < time.Second {
				mu.Unlock()
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
		}
		lastRequest[ip] = now
		mu.Unlock()

		// Get backend per request
		backend := getLeastConnBackend()
		if backend == nil {
			http.Error(w, "No backends available", http.StatusServiceUnavailable)
			return
		}
		targetURL, err := url.Parse(backend.URL)
		if err != nil {
			http.Error(w, "Invalid backend URL", http.StatusBadRequest)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		log.Printf("%s %s %s -> %s", ip, r.Method, r.URL.Path, targetURL)
		//free up conn count after req is served

		defer func() {
			lcMu.Lock()
			backend.ActiveCon--
			lcMu.Unlock()
			log.Printf("Backend chosen: %s (Active: %d)", backend.URL, backend.ActiveCon)
		}()

		proxy.ServeHTTP(w, r)

	})

	log.Println("Starting reverse proxy server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
