package main

import (
	"fmt"
	"net/http"
)

func Server2() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello from backend server 2")
	})
	http.ListenAndServe(":9001", mux)
}
