package main

import (
	"fmt"
	"net/http"
	"time"
)

func Server1() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprintln(w, "hello from backend server 1")
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		fmt.Fprintln(w, "Server 1 is healthy !")
	})
	http.ListenAndServe(":9000", mux)
}
