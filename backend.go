package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(w, "hello from backend server")
	})
	http.ListenAndServe(":9000", nil)
}
