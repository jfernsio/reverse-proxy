package main

import (
	"fmt"
	"net/http"
	"time"
)

func Server2() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello from backend server 2")
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		fmt.Fprintln(w, "Server 2 is healthy! ")
	})
	http.ListenAndServe(":9001", mux)
}
