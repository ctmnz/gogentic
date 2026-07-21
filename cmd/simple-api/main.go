package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World"
	}
	start := time.Now()
	fmt.Fprintf(w, "Hello, %s", name)
	log.Printf("[API] GET /hello?name=%s -> Hello, %s (%v)", name, name, time.Since(start))
}

func main() {
	http.HandleFunc("/hello", helloHandler)
	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
