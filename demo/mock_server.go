package main

// фейковый сервер для тестов балансировщика запросов

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	var port string
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	if port == "" {
		log.Fatal("PORT environment variable is required")
	}

	var countRequests uint64
	// Простой HTTP-сервер
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from backend server on port %s\n", port)
		countRequests += 1
		log.Println(countRequests) // вывожу кол-во запросов на сервер, для проверки работы балансировщика
	})

	// Эндпоинт для Health Checks
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("Backend server started on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
