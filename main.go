package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var lastRequest = make(map[string]time.Time)

func main() {
	target, err := url.Parse("http://localhost:9000")

	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		now := time.Now()

		if t, ok := lastRequest[ip]; ok {
			if now.Sub(t) < time.Second {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
		}
		lastRequest[ip] = now
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	log.Println("Starting reverse proxy server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
