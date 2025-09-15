package main

import (
	"log"
	"net/http"

	"github.com/n-korel/nexus-drive-go/shared/env"
)

var (
	httpAddr = env.GetString("HTTP_ADDR", ":8081")
)

func main() {
	log.Println("API Gateway")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from API Gateway"))
	})

	http.ListenAndServe(httpAddr, nil)
}